package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	StatusRequested TripStatus = "solicitado"
	StatusApproved  TripStatus = "aprovado"
	StatusCanceled  TripStatus = "cancelado"
)

func (s TripStatus) IsValid() bool {
	switch s {
	case StatusRequested, StatusApproved, StatusCanceled:
		return true
	}
	return false
}

type Trip struct {
	ID          uuid.UUID  `json:"id"`
	RequesterID uuid.UUID  `json:"requester_id"`
	Destination string     `json:"destination"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	Status      TripStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Validate checks if the trip data is valid according to business rules
func (t *Trip) Validate() error {
	// Check required fields
	if t.RequesterID == uuid.Nil {
		return errors.New("requester_id is required")
	}

	if t.Destination == "" {
		return errors.New("destination is required")
	}

	// Check if StartDate is zero
	if t.StartDate.IsZero() {
		return errors.New("start_date is required")
	}

	// Check if EndDate is zero
	if t.EndDate.IsZero() {
		return errors.New("end_date is required")
	}

	// Check if EndDate is after StartDate
	if !t.EndDate.After(t.StartDate) {
		return errors.New("end_date must be after start_date")
	}

	// Check if Status is valid
	if !t.Status.IsValid() {
		return errors.New("invalid status")
	}

	return nil
}

type ListTripsParams struct {
	RequesterID *uuid.UUID
	Status      *TripStatus
	Destination *string
	StartDate   *time.Time
	EndDate     *time.Time
}

type TripRepository interface {
	Create(ctx context.Context, trip *Trip) error
	FindByID(ctx context.Context, id uuid.UUID) (*Trip, error)
	List(ctx context.Context, params ListTripsParams) ([]*Trip, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status TripStatus) error
}
