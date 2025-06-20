package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
)

var (
	ErrTripNotFound     = errors.New("trip not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrSelfApproval     = errors.New("you cannot approve or cancel your own trip")
	ErrInvalidStatus    = errors.New("invalid status for this operation")
	ErrCancelNotAllowed = errors.New("cannot cancel a trip that starts in 7 days or less")
)

type TripService struct {
	tripRepo domain.TripRepository
	userRepo domain.UserRepository // Needed to fetch user for notifications
	notifier NotificationService
}

func NewTripService(tripRepo domain.TripRepository, userRepo domain.UserRepository, notifier NotificationService) *TripService {
	return &TripService{
		tripRepo: tripRepo,
		userRepo: userRepo,
		notifier: notifier,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, requesterID uuid.UUID, dest string, start, end time.Time) (*domain.Trip, error) {
	trip := &domain.Trip{
		ID:          uuid.New(),
		RequesterID: requesterID,
		Destination: dest,
		StartDate:   start,
		EndDate:     end,
		Status:      domain.StatusRequested,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate trip before saving
	if err := trip.Validate(); err != nil {
		return nil, err
	}

	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *TripService) GetTripByID(ctx context.Context, tripID, userID uuid.UUID) (*domain.Trip, error) {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, ErrTripNotFound
	}
	// Allow any user to see any trip to enable approval by other users
	return trip, nil
}

func (s *TripService) ListTrips(ctx context.Context, params domain.ListTripsParams) ([]*domain.Trip, error) {
	// The repository will be filtered by requesterID, so it's secure.
	return s.tripRepo.List(ctx, params)
}

func (s *TripService) UpdateTripStatus(ctx context.Context, tripID, updaterID uuid.UUID, newStatus domain.TripStatus) error {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return ErrTripNotFound
	}

	// Rule: The user who created the request cannot change its status.
	if trip.RequesterID == updaterID {
		return ErrSelfApproval
	}

	if err := s.tripRepo.UpdateStatus(ctx, tripID, newStatus); err != nil {
		return err
	}

	// Send notification
	requester, err := s.userRepo.FindByID(ctx, trip.RequesterID)
	if err == nil && requester != nil {
		message := fmt.Sprintf("Your trip to %s has been %s.", trip.Destination, newStatus)
		s.notifier.Send(requester, trip, message)
	}

	return nil
}

func (s *TripService) CancelApprovedTrip(ctx context.Context, tripID, cancelingUserID uuid.UUID) error {
	trip, err := s.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return ErrTripNotFound
	}

	// Rule: Only the requester can cancel their own approved trip.
	if trip.RequesterID != cancelingUserID {
		return ErrPermissionDenied
	}

	if trip.Status != domain.StatusApproved {
		return ErrInvalidStatus
	}

	// Business Rule: Cannot cancel a trip that starts within the next 7 days.
	if time.Until(trip.StartDate) < 7*24*time.Hour {
		return ErrCancelNotAllowed
	}

	if err := s.tripRepo.UpdateStatus(ctx, tripID, domain.StatusCanceled); err != nil {
		return err
	}

	// Send notification to the requester
	requester, err := s.userRepo.FindByID(ctx, trip.RequesterID)
	if err == nil && requester != nil {
		message := fmt.Sprintf("Your trip to %s has been canceled.", trip.Destination)
		s.notifier.Send(requester, trip, message)
	}

	return nil
}
