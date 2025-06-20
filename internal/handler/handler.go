package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/service"
	"strings"
)

// Handler holds all services that the handlers will need.
type Handler struct {
	userService  *service.UserService
	tripService  *service.TripService
	validate     *validator.Validate
	jwtSecretKey string
}

func NewHandler(userSvc *service.UserService, tripSvc *service.TripService, jwtSecretKey string) *Handler {
	return &Handler{
		userService:  userSvc,
		tripService:  tripSvc,
		validate:     validator.New(),
		jwtSecretKey: jwtSecretKey,
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

// parseValidationErrors converts Gin's validation errors to our custom ValidationErrors format
func parseValidationErrors(err error) *domain.ValidationErrors {
	validationErrors := domain.NewValidationErrors()

	// Check if the error is a validator.ValidationErrors type
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			fieldName := strings.ToLower(fieldErr.Field())

			// Map validation tags to user-friendly error messages
			switch fieldErr.Tag() {
			case "required":
				validationErrors.Add(fieldName + " is required")
			case "email":
				validationErrors.Add("invalid email format")
			case "min":
				validationErrors.Add(fieldName + " must be at least " + fieldErr.Param() + " characters")
			default:
				validationErrors.Add(fieldName + " is invalid")
			}
		}
	} else {
		// If it's not a validation error, add the error message as is
		validationErrors.Add(err.Error())
	}

	return validationErrors
}
