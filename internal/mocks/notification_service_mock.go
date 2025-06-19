package mocks

import (
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService is a mock implementation of service.NotificationService
type MockNotificationService struct {
	mock.Mock
}

// Send mocks the Send method
func (m *MockNotificationService) Send(user *domain.User, trip *domain.Trip, message string) {
	m.Called(user, trip, message)
}
