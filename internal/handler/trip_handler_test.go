package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/handler"
	"github.com/jimmmmisss/api-viagens/internal/mocks"
	"github.com/jimmmmisss/api-viagens/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Setup test router with mock repositories and a fixed userID
func setupTripTestRouter() (*gin.Engine, *mocks.MockTripRepository, *mocks.MockUserRepository, *mocks.MockNotificationService, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	
	mockTripRepo := new(mocks.MockTripRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockNotifier := new(mocks.MockNotificationService)
	
	userService := service.NewUserService(mockUserRepo)
	tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)
	
	h := handler.NewHandler(userService, tripService)
	
	// Create a fixed userID for testing
	userID := uuid.New()
	
	// Add middleware to set userID in context for authenticated routes
	authMiddleware := func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
	
	// Setup routes
	tripRoutes := router.Group("/")
	tripRoutes.Use(authMiddleware)
	{
		tripRoutes.POST("/trips", h.CreateTrip)
		tripRoutes.GET("/trips/:id", h.GetTripByID)
		tripRoutes.PATCH("/trips/:id/status", h.UpdateTripStatus)
		tripRoutes.POST("/trips/:id/cancel", h.CancelApprovedTrip)
	}
	
	return router, mockTripRepo, mockUserRepo, mockNotifier, userID
}

func TestCreateTrip(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, _ := setupTripTestRouter()
		
		startDate := time.Now().AddDate(0, 1, 0) // 1 month from now
		endDate := startDate.AddDate(0, 0, 7)    // 7 days after start
		
		// Mock behavior
		mockTripRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trip")).Return(nil)
		
		// Create request
		reqBody := map[string]interface{}{
			"destination": "Paris",
			"start_date":  startDate.Format(time.RFC3339),
			"end_date":    endDate.Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response domain.Trip
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Paris", response.Destination)
		assert.Equal(t, domain.StatusRequested, response.Status)
		
		mockTripRepo.AssertExpectations(t)
	})
	
	t.Run("Invalid request body", func(t *testing.T) {
		// Arrange
		router, _, _, _, _ := setupTripTestRouter()
		
		// Create request with invalid body
		reqBody := map[string]interface{}{
			// Missing destination
			"start_date": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
			"end_date":   time.Now().AddDate(0, 1, 7).Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Invalid date range", func(t *testing.T) {
		// Arrange
		router, _, _, _, _ := setupTripTestRouter()
		
		endDate := time.Now().AddDate(0, 1, 0)
		startDate := endDate.AddDate(0, 0, 7) // Start date after end date
		
		// Create request
		reqBody := map[string]interface{}{
			"destination": "Paris",
			"start_date":  startDate.Format(time.RFC3339),
			"end_date":    endDate.Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Database error", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, _ := setupTripTestRouter()
		
		startDate := time.Now().AddDate(0, 1, 0)
		endDate := startDate.AddDate(0, 0, 7)
		dbError := errors.New("database error")
		
		// Mock behavior
		mockTripRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Trip")).Return(dbError)
		
		// Create request
		reqBody := map[string]interface{}{
			"destination": "Paris",
			"start_date":  startDate.Format(time.RFC3339),
			"end_date":    endDate.Format(time.RFC3339),
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/trips", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
}

func TestUpdateTripStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, mockUserRepo, mockNotifier, _ := setupTripTestRouter()
		
		tripID := uuid.New()
		requesterID := uuid.New()
		
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
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(trip, nil)
		mockTripRepo.On("UpdateStatus", mock.Anything, tripID, domain.StatusApproved).Return(nil)
		mockUserRepo.On("FindByID", mock.Anything, requesterID).Return(user, nil)
		mockNotifier.On("Send", user, trip, mock.AnythingOfType("string")).Return()
		
		// Create request
		reqBody := map[string]interface{}{
			"status": "aprovado",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/trips/%s/status", tripID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockTripRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockNotifier.AssertExpectations(t)
	})
	
	t.Run("Invalid trip ID", func(t *testing.T) {
		// Arrange
		router, _, _, _, _ := setupTripTestRouter()
		
		// Create request
		reqBody := map[string]interface{}{
			"status": "aprovado",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", "/trips/invalid-uuid/status", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Invalid status value", func(t *testing.T) {
		// Arrange
		router, _, _, _, _ := setupTripTestRouter()
		
		tripID := uuid.New()
		
		// Create request
		reqBody := map[string]interface{}{
			"status": "invalid_status",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/trips/%s/status", tripID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Self approval not allowed", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, userID := setupTripTestRouter()
		
		tripID := uuid.New()
		
		// Mock behavior - we'll set up the trip to have the same requesterID as the userID in the context
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(&domain.Trip{
			ID:          tripID,
			RequesterID: userID, // Same as the userID in the context
			Status:      domain.StatusRequested,
		}, nil)
		
		// Create request
		reqBody := map[string]interface{}{
			"status": "aprovado",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/trips/%s/status", tripID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
	
	t.Run("Trip not found", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, _ := setupTripTestRouter()
		
		tripID := uuid.New()
		
		// Mock behavior
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(nil, nil)
		
		// Create request
		reqBody := map[string]interface{}{
			"status": "aprovado",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/trips/%s/status", tripID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
}

func TestCancelApprovedTrip(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, userID := setupTripTestRouter()
		
		tripID := uuid.New()
		startDate := time.Now().AddDate(0, 1, 0) // 1 month from now (more than 7 days)
		
		// Mock behavior - we'll set up the trip to have the same requesterID as the userID in the context
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(&domain.Trip{
			ID:          tripID,
			RequesterID: userID, // Same as the userID in the context
			Status:      domain.StatusApproved,
			StartDate:   startDate,
		}, nil)
		mockTripRepo.On("UpdateStatus", mock.Anything, tripID, domain.StatusCanceled).Return(nil)
		
		// Create request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/trips/%s/cancel", tripID), nil)
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
	
	t.Run("Invalid trip ID", func(t *testing.T) {
		// Arrange
		router, _, _, _, _ := setupTripTestRouter()
		
		// Create request
		req, _ := http.NewRequest("POST", "/trips/invalid-uuid/cancel", nil)
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Trip not found", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, _ := setupTripTestRouter()
		
		tripID := uuid.New()
		
		// Mock behavior
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(nil, nil)
		
		// Create request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/trips/%s/cancel", tripID), nil)
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
	
	t.Run("Permission denied", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, _ := setupTripTestRouter()
		
		tripID := uuid.New()
		requesterID := uuid.New() // Different from the userID in the context
		
		// Mock behavior
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(&domain.Trip{
			ID:          tripID,
			RequesterID: requesterID,
			Status:      domain.StatusApproved,
		}, nil)
		
		// Create request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/trips/%s/cancel", tripID), nil)
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
	
	t.Run("Invalid status", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, userID := setupTripTestRouter()
		
		tripID := uuid.New()
		
		// Mock behavior - we'll set up the trip to have the same requesterID as the userID in the context
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(&domain.Trip{
			ID:          tripID,
			RequesterID: userID, // Same as the userID in the context
			Status:      domain.StatusRequested, // Not approved
		}, nil)
		
		// Create request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/trips/%s/cancel", tripID), nil)
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
	
	t.Run("Cancel not allowed within 7 days", func(t *testing.T) {
		// Arrange
		router, mockTripRepo, _, _, userID := setupTripTestRouter()
		
		tripID := uuid.New()
		startDate := time.Now().AddDate(0, 0, 3) // Only 3 days from now
		
		// Mock behavior - we'll set up the trip to have the same requesterID as the userID in the context
		mockTripRepo.On("FindByID", mock.Anything, tripID).Return(&domain.Trip{
			ID:          tripID,
			RequesterID: userID, // Same as the userID in the context
			Status:      domain.StatusApproved,
			StartDate:   startDate,
		}, nil)
		
		// Create request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/trips/%s/cancel", tripID), nil)
		
		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Assert
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockTripRepo.AssertExpectations(t)
	})
}