package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/handler"
	"github.com/jimmmmisss/api-viagens/internal/mocks"
	"github.com/jimmmmisss/api-viagens/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Setup test router with mock repositories
func setupTestRouter() (*gin.Engine, *mocks.MockUserRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	mockUserRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockUserRepo)

	// We don't need to mock the trip service for user handler tests
	mockTripRepo := new(mocks.MockTripRepository)
	mockNotifier := new(mocks.MockNotificationService)
	tripService := service.NewTripService(mockTripRepo, mockUserRepo, mockNotifier)

	h := handler.NewHandler(userService, tripService)

	router.POST("/register", h.RegisterUser)
	router.POST("/login", h.LoginUser)

	return router, mockUserRepo
}

func TestRegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		userID := uuid.New()
		user := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, nil)
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
			createdUser := args.Get(1).(*domain.User)
			// Copy the ID to the user we'll return
			user.ID = createdUser.ID
		})

		// Create request
		reqBody := map[string]interface{}{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Test User", response["name"])
		assert.Equal(t, "test@example.com", response["email"])

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		// Arrange
		router, _ := setupTestRouter()

		// Create request with invalid body
		reqBody := map[string]interface{}{
			"name":     "Test User",
			// Missing email
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("User already exists", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		existingUser := &domain.User{
			ID:    uuid.New(),
			Email: "existing@example.com",
		}

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

		// Create request
		reqBody := map[string]interface{}{
			"name":     "Existing User",
			"email":    "existing@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusConflict, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		dbError := errors.New("database error")

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, dbError)

		// Create request
		reqBody := map[string]interface{}{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestLoginUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		hashedPassword := "$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8iw0hLyhsiG" // hashed "password123"
		userID := uuid.New()
		user := &domain.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

		// Create request
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "token")
		assert.NotEmpty(t, response["token"])

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		// Arrange
		router, _ := setupTestRouter()

		// Create request with invalid body
		reqBody := map[string]interface{}{
			// Missing email
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid credentials - user not found", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)

		// Create request
		reqBody := map[string]interface{}{
			"email":    "nonexistent@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid credentials - wrong password", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		hashedPassword := "$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8iw0hLyhsiG" // hashed "password123"
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
		}

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

		// Create request
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange
		router, mockUserRepo := setupTestRouter()

		dbError := errors.New("database error")

		// Mock behavior
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, dbError)

		// Create request
		reqBody := map[string]interface{}{
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUserRepo.AssertExpectations(t)
	})
}
