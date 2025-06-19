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
