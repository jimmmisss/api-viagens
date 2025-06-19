package service

import (
	"log"

	"github.com/jimmmmisss/api-viagens/internal/domain"
)

// NotificationService defines the interface for sending notifications.
type NotificationService interface {
	Send(user *domain.User, trip *domain.Trip, message string)
}

// logNotificationService is a simple implementation that logs to the console.
type logNotificationService struct{}

func NewLogNotificationService() NotificationService {
	return &logNotificationService{}
}

// Send simulates sending a notification by printing to the console.
func (s *logNotificationService) Send(user *domain.User, trip *domain.Trip, message string) {
	log.Printf("--- NOTIFICATION ---")
	log.Printf("To: %s (%s)", user.Name, user.Email)
	log.Printf("Trip ID: %s", trip.ID)
	log.Printf("Destination: %s", trip.Destination)
	log.Printf("Message: %s", message)
	log.Printf("--- END NOTIFICATION ---")
}
