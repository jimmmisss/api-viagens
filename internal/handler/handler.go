package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/service"
)

// Handler holds all services that the handlers will need.
type Handler struct {
	userService *service.UserService
	tripService *service.TripService
	validate    *validator.Validate
}

func NewHandler(userSvc *service.UserService, tripSvc *service.TripService) *Handler {
	return &Handler{
		userService: userSvc,
		tripService: tripSvc,
		validate:    validator.New(),
	}
}

// Helper to get userID from context
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}
