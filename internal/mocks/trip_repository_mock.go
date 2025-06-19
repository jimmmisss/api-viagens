package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockTripRepository is a mock implementation of domain.TripRepository
type MockTripRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (m *MockTripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	args := m.Called(ctx, trip)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockTripRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Trip), args.Error(1)
}

// List mocks the List method
func (m *MockTripRepository) List(ctx context.Context, params domain.ListTripsParams) ([]*domain.Trip, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Trip), args.Error(1)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockTripRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TripStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}
