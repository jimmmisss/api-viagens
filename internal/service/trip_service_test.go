package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/mocks"
	"github.com/jimmmmisss/api-viagens/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTripService_CreateTrip(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange - create new mocks for this test case
		mockTripRepo := new(mocks.MockTripRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockNotifier := new(mocks.MockNotificationService)
		tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)

		requesterID := uuid.New()
		destination := "Paris"
		startDate := time.Now().AddDate(0, 1, 0) // 1 month from now
		endDate := startDate.AddDate(0, 0, 7)    // 7 days after start

		// Mock behavior
		mockTripRepo.On("Create", ctx, mock.AnythingOfType("*domain.Trip")).Return(nil)

		// Act
		trip, err := tripService.CreateTrip(ctx, requesterID, destination, startDate, endDate)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, trip)
		assert.Equal(t, requesterID, trip.RequesterID)
		assert.Equal(t, destination, trip.Destination)
		assert.Equal(t, startDate, trip.StartDate)
		assert.Equal(t, endDate, trip.EndDate)
		assert.Equal(t, domain.StatusRequested, trip.Status)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange - create new mocks for this test case
		mockTripRepo := new(mocks.MockTripRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockNotifier := new(mocks.MockNotificationService)
		tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)

		requesterID := uuid.New()
		destination := "Paris"
		startDate := time.Now().AddDate(0, 1, 0)
		endDate := startDate.AddDate(0, 0, 7)
		dbError := errors.New("database error")

		// Mock behavior - use AnythingOfType to match any Trip object
		mockTripRepo.On("Create", ctx, mock.AnythingOfType("*domain.Trip")).Return(dbError)

		// Act
		trip, err := tripService.CreateTrip(ctx, requesterID, destination, startDate, endDate)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, trip)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Validation error - end date before start date", func(t *testing.T) {
		// Arrange - create new mocks for this test case
		mockTripRepo := new(mocks.MockTripRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockNotifier := new(mocks.MockNotificationService)
		tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)

		requesterID := uuid.New()
		destination := "Paris"
		startDate := time.Now().AddDate(0, 1, 0)
		endDate := startDate.AddDate(0, 0, -1) // Invalid: end date before start date

		// No mock behavior needed as validation should fail before repository is called

		// Act
		trip, err := tripService.CreateTrip(ctx, requesterID, destination, startDate, endDate)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end_date must be after start_date")
		assert.Nil(t, trip)
		// Verify that Create was never called
		mockTripRepo.AssertNotCalled(t, "Create")
	})
}

func TestTripService_GetTripByID(t *testing.T) {
	// Arrange
	mockTripRepo := new(mocks.MockTripRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockNotifier := new(mocks.MockNotificationService)
	tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		userID := uuid.New() // Same as requesterID
		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: userID,
			Destination: "Paris",
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)

		// Act
		foundTrip, err := tripService.GetTripByID(ctx, tripID, userID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, trip, foundTrip)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Trip not found", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		userID := uuid.New()

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(nil, nil)

		// Act
		trip, err := tripService.GetTripByID(ctx, tripID, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrTripNotFound, err)
		assert.Nil(t, trip)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Permission denied", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		userID := uuid.New() // Different from requesterID
		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Destination: "Paris",
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)

		// Act
		foundTrip, err := tripService.GetTripByID(ctx, tripID, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrPermissionDenied, err)
		assert.Nil(t, foundTrip)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		userID := uuid.New()
		dbError := errors.New("database error")

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(nil, dbError)

		// Act
		trip, err := tripService.GetTripByID(ctx, tripID, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, trip)
		mockTripRepo.AssertExpectations(t)
	})
}

func TestTripService_UpdateTripStatus(t *testing.T) {
	// Arrange
	mockTripRepo := new(mocks.MockTripRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockNotifier := new(mocks.MockNotificationService)
	tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		updaterID := uuid.New() // Different from requester
		newStatus := domain.StatusApproved

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusRequested,
			Destination: "Paris",
		}

		user := &domain.User{
			ID:    requesterID,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)
		mockTripRepo.On("UpdateStatus", ctx, tripID, newStatus).Return(nil)
		mockUserRepo.On("FindByID", ctx, requesterID).Return(user, nil)
		mockNotifier.On("Send", user, trip, mock.AnythingOfType("string")).Return()

		// Act
		err := tripService.UpdateTripStatus(ctx, tripID, updaterID, newStatus)

		// Assert
		assert.NoError(t, err)
		mockTripRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockNotifier.AssertExpectations(t)
	})

	t.Run("Trip not found", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		updaterID := uuid.New()
		newStatus := domain.StatusApproved

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(nil, nil)

		// Act
		err := tripService.UpdateTripStatus(ctx, tripID, updaterID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrTripNotFound, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Self approval not allowed", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		newStatus := domain.StatusApproved

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusRequested,
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)

		// Act
		err := tripService.UpdateTripStatus(ctx, tripID, requesterID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrSelfApproval, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Database error on FindByID", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		updaterID := uuid.New()
		newStatus := domain.StatusApproved
		dbError := errors.New("database error")

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(nil, dbError)

		// Act
		err := tripService.UpdateTripStatus(ctx, tripID, updaterID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Database error on UpdateStatus", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		updaterID := uuid.New() // Different from requester
		newStatus := domain.StatusApproved
		dbError := errors.New("database error")

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusRequested,
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)
		mockTripRepo.On("UpdateStatus", ctx, tripID, newStatus).Return(dbError)

		// Act
		err := tripService.UpdateTripStatus(ctx, tripID, updaterID, newStatus)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockTripRepo.AssertExpectations(t)
	})
}

func TestTripService_CancelApprovedTrip(t *testing.T) {
	// Arrange
	mockTripRepo := new(mocks.MockTripRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockNotifier := new(mocks.MockNotificationService)
	tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		startDate := time.Now().AddDate(0, 1, 0) // 1 month from now (more than 7 days)

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusApproved,
			StartDate:   startDate,
			Destination: "Paris",
		}

		user := &domain.User{
			ID:    requesterID,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)
		mockTripRepo.On("UpdateStatus", ctx, tripID, domain.StatusCanceled).Return(nil)
		mockUserRepo.On("FindByID", ctx, requesterID).Return(user, nil)
		mockNotifier.On("Send", user, trip, mock.AnythingOfType("string")).Return()

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, requesterID)

		// Assert
		assert.NoError(t, err)
		mockTripRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockNotifier.AssertExpectations(t)
	})

	t.Run("Trip not found", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(nil, nil)

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, requesterID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrTripNotFound, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Permission denied", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		cancelingUserID := uuid.New() // Different from requester

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusApproved,
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, cancelingUserID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrPermissionDenied, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Invalid status", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusRequested, // Not approved
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, requesterID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidStatus, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Cancel not allowed within 7 days", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		startDate := time.Now().AddDate(0, 0, 3) // Only 3 days from now

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusApproved,
			StartDate:   startDate,
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, requesterID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrCancelNotAllowed, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Database error on FindByID", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		dbError := errors.New("database error")

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(nil, dbError)

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, requesterID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockTripRepo.AssertExpectations(t)
	})

	t.Run("Database error on UpdateStatus", func(t *testing.T) {
		// Arrange
		tripID := uuid.New()
		requesterID := uuid.New()
		startDate := time.Now().AddDate(0, 1, 0) // 1 month from now
		dbError := errors.New("database error")

		trip := &domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusApproved,
			StartDate:   startDate,
		}

		// Mock behavior
		mockTripRepo.On("FindByID", ctx, tripID).Return(trip, nil)
		mockTripRepo.On("UpdateStatus", ctx, tripID, domain.StatusCanceled).Return(dbError)

		// Act
		err := tripService.CancelApprovedTrip(ctx, tripID, requesterID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockTripRepo.AssertExpectations(t)
	})
}
