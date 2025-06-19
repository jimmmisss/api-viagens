package domain

import (
	"context"
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
	validationErrors := NewValidationErrors()

	// Check required fields
	validationErrors.AddIf(t.RequesterID == uuid.Nil, "requester_id is required")
	validationErrors.AddIf(t.Destination == "", "destination is required")

	// Check if StartDate is zero
	validationErrors.AddIf(t.StartDate.IsZero(), "start_date is required")

	// Check if EndDate is zero
	validationErrors.AddIf(t.EndDate.IsZero(), "end_date is required")

	// Check if EndDate is after StartDate
	if !t.StartDate.IsZero() && !t.EndDate.IsZero() {
		validationErrors.AddIf(!t.EndDate.After(t.StartDate), "end_date must be after start_date")
	}

	// Check if Status is valid
	validationErrors.AddIf(!t.Status.IsValid(), "invalid status")

	if validationErrors.HasErrors() {
		return validationErrors
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
