package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTrip_Validate(t *testing.T) {
	// Setup valid trip for reuse
	validRequesterID := uuid.New()
	validStartDate := time.Now().Add(24 * time.Hour)   // tomorrow
	validEndDate := validStartDate.Add(48 * time.Hour) // 2 days after start

	t.Run("Valid trip", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: validRequesterID,
			Destination: "Paris",
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			Status:      StatusRequested,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := trip.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing requester ID", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: uuid.Nil, // Invalid: zero value
			Destination: "Paris",
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			Status:      StatusRequested,
		}

		err := trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requester_id is required")
	})

	t.Run("Missing destination", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: validRequesterID,
			Destination: "", // Invalid: empty string
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			Status:      StatusRequested,
		}

		err := trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination is required")
	})

	t.Run("Missing start date", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: validRequesterID,
			Destination: "Paris",
			StartDate:   time.Time{}, // Invalid: zero value
			EndDate:     validEndDate,
			Status:      StatusRequested,
		}

		err := trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start_date is required")
	})

	t.Run("Missing end date", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: validRequesterID,
			Destination: "Paris",
			StartDate:   validStartDate,
			EndDate:     time.Time{}, // Invalid: zero value
			Status:      StatusRequested,
		}

		err := trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end_date is required")
	})

	t.Run("End date not after start date", func(t *testing.T) {
		// Test with end date equal to start date
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: validRequesterID,
			Destination: "Paris",
			StartDate:   validStartDate,
			EndDate:     validStartDate, // Invalid: same as start date
			Status:      StatusRequested,
		}

		err := trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end_date must be after start_date")

		// Test with end date before start date
		trip.EndDate = validStartDate.Add(-24 * time.Hour) // Invalid: before start date
		err = trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end_date must be after start_date")
	})

	t.Run("Invalid status", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: validRequesterID,
			Destination: "Paris",
			StartDate:   validStartDate,
			EndDate:     validEndDate,
			Status:      "invalid_status", // Invalid status
		}

		err := trip.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("Multiple validation errors", func(t *testing.T) {
		trip := &Trip{
			ID:          uuid.New(),
			RequesterID: uuid.Nil,         // Invalid: zero value
			Destination: "",               // Invalid: empty string
			StartDate:   time.Time{},      // Invalid: zero value
			EndDate:     time.Time{},      // Invalid: zero value
			Status:      "invalid_status", // Invalid status
		}

		err := trip.Validate()
		assert.Error(t, err)

		// Check that all errors are reported
		errMsg := err.Error()
		assert.Contains(t, errMsg, "requester_id is required")
		assert.Contains(t, errMsg, "destination is required")
		assert.Contains(t, errMsg, "start_date is required")
		assert.Contains(t, errMsg, "end_date is required")
		assert.Contains(t, errMsg, "invalid status")

		// Check that it's a ValidationErrors type
		validationErrs, ok := err.(*ValidationErrors)
		assert.True(t, ok, "Error should be of type *ValidationErrors")
		assert.Equal(t, 5, len(validationErrs.GetErrors()))
	})
}
