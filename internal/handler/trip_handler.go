package handler

import (
	"errors"
	gin "github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/service"
)

type createTripRequest struct {
	Destination string    `json:"destination" binding:"required"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
}

func (h *Handler) CreateTrip(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	var req createTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := parseValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors.GetErrors()})
		return
	}

	if req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be after start_date"})
		return
	}

	trip, err := h.tripService.CreateTrip(c.Request.Context(), userID, req.Destination, req.StartDate, req.EndDate)
	if err != nil {
		// Check if it's a validation error
		if validationErrs, ok := err.(*domain.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trip"})
		return
	}

	c.JSON(http.StatusCreated, trip)
}

func (h *Handler) GetTripByID(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID format"})
		return
	}

	trip, err := h.tripService.GetTripByID(c.Request.Context(), tripID, userID)
	if err != nil {
		if errors.Is(err, service.ErrTripNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrPermissionDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		// Check if it's a validation error
		if validationErrs, ok := err.(*domain.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve trip"})
		return
	}

	c.JSON(http.StatusOK, trip)
}

func (h *Handler) ListTrips(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	params := domain.ListTripsParams{}

	// If requester_id query param is provided, filter by that requester
	// Otherwise, show all trips to allow other users to approve them
	if requesterID := c.Query("requester_id"); requesterID != "" {
		if id, err := uuid.Parse(requesterID); err == nil {
			params.RequesterID = &id
		}
	} else if onlyMine := c.Query("only_mine"); onlyMine == "true" {
		// If only_mine=true is specified, show only the current user's trips
		params.RequesterID = &userID
	}

	if status := c.Query("status"); status != "" {
		s := domain.TripStatus(status)
		if s.IsValid() {
			params.Status = &s
		}
	}
	if dest := c.Query("destination"); dest != "" {
		params.Destination = &dest
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			params.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			params.EndDate = &t
		}
	}

	trips, err := h.tripService.ListTrips(c.Request.Context(), params)
	if err != nil {
		// Check if it's a validation error
		if validationErrs, ok := err.(*domain.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list trips"})
		return
	}

	c.JSON(http.StatusOK, trips)
}

type updateStatusRequest struct {
	Status domain.TripStatus `json:"status" binding:"required"`
}

func (h *Handler) UpdateTripStatus(c *gin.Context) {
	updaterID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID format"})
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := parseValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors.GetErrors()})
		return
	}
	if !req.Status.IsValid() || req.Status == domain.StatusRequested {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	err = h.tripService.UpdateTripStatus(c.Request.Context(), tripID, updaterID, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTripNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrSelfApproval):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			// Check if it's a validation error
			if validationErrs, ok := err.(*domain.ValidationErrors); ok {
				c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trip status"})
			}
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trip status updated successfully"})
}

func (h *Handler) CancelApprovedTrip(c *gin.Context) {
	cancelingUserID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID format"})
		return
	}

	err = h.tripService.CancelApprovedTrip(c.Request.Context(), tripID, cancelingUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTripNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrPermissionDenied), errors.Is(err, service.ErrCancelNotAllowed), errors.Is(err, service.ErrInvalidStatus):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			// Check if it's a validation error
			if validationErrs, ok := err.(*domain.ValidationErrors); ok {
				c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel trip"})
			}
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trip cancellation successful"})
}
