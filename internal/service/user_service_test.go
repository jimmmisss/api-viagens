package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/mocks"
	"github.com/jimmmmisss/api-viagens/internal/service"
	"github.com/jimmmmisss/api-viagens/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		name := "Test User"
		email := "test@example.com"
		password := "password123"

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

		// Act
		user, err := userService.Register(ctx, name, email, password)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, email, user.Email)
		assert.NotEmpty(t, user.PasswordHash)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User already exists", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		name := "Existing User"
		email := "existing@example.com"
		password := "password123"
		existingUser := &domain.User{
			ID:    uuid.New(),
			Email: email,
		}

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(existingUser, nil)

		// Act
		user, err := userService.Register(ctx, name, email, password)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrUserAlreadyExists, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		name := "Test User"
		email := "test@example.com"
		password := "password123"
		dbError := errors.New("database error")

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(nil, dbError)

		// Act
		user, err := userService.Register(ctx, name, email, password)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		email := "test@example.com"
		password := "password123"

		// Generate a valid hash for the password
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		user := &domain.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: hashedPassword,
		}

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(user, nil)

		// Act
		loggedInUser, err := userService.Login(ctx, email, password)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, user, loggedInUser)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		email := "nonexistent@example.com"
		password := "password123"

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(nil, nil)

		// Act
		user, err := userService.Login(ctx, email, password)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidCredentials, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid password", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		email := "test@example.com"
		password := "wrongpassword"
		correctPassword := "password123"

		// Generate a valid hash for the correct password
		hashedPassword, err := utils.HashPassword(correctPassword)
		assert.NoError(t, err)

		user := &domain.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: hashedPassword,
		}

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(user, nil)

		// Act
		loggedInUser, err := userService.Login(ctx, email, password)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidCredentials, err)
		assert.Nil(t, loggedInUser)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange - create new mock for this test case
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		email := "test@example.com"
		password := "password123"
		dbError := errors.New("database error")

		// Mock behavior
		mockRepo.On("FindByEmail", ctx, email).Return(nil, dbError)

		// Act
		user, err := userService.Login(ctx, email, password)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	// Arrange
	mockRepo := new(mocks.MockUserRepository)
	userService := service.NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		user := &domain.User{
			ID:    userID,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// Mock behavior
		mockRepo.On("FindByID", ctx, userID).Return(user, nil)

		// Act
		foundUser, err := userService.GetUserByID(ctx, userID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, user, foundUser)
		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		// Arrange
		userID := uuid.New()

		// Mock behavior
		mockRepo.On("FindByID", ctx, userID).Return(nil, nil)

		// Act
		user, err := userService.GetUserByID(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, service.ErrUserNotFound, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		dbError := errors.New("database error")

		// Mock behavior
		mockRepo.On("FindByID", ctx, userID).Return(nil, dbError)

		// Act
		user, err := userService.GetUserByID(ctx, userID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}
