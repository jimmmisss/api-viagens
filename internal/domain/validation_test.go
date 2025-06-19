package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationErrors(t *testing.T) {
	t.Run("Empty validation errors", func(t *testing.T) {
		validationErrors := NewValidationErrors()

		assert.False(t, validationErrors.HasErrors())
		assert.Empty(t, validationErrors.GetErrors())
		assert.Empty(t, validationErrors.Error())
	})

	t.Run("Add single error", func(t *testing.T) {
		validationErrors := NewValidationErrors()
		validationErrors.Add("test error")

		assert.True(t, validationErrors.HasErrors())
		assert.Len(t, validationErrors.GetErrors(), 1)
		assert.Equal(t, "test error", validationErrors.GetErrors()[0])
		assert.Equal(t, "Validation errors: test error", validationErrors.Error())
	})

	t.Run("Add multiple errors", func(t *testing.T) {
		validationErrors := NewValidationErrors()
		validationErrors.Add("error 1")
		validationErrors.Add("error 2")
		validationErrors.Add("error 3")

		assert.True(t, validationErrors.HasErrors())
		assert.Len(t, validationErrors.GetErrors(), 3)
		assert.Equal(t, "error 1", validationErrors.GetErrors()[0])
		assert.Equal(t, "error 2", validationErrors.GetErrors()[1])
		assert.Equal(t, "error 3", validationErrors.GetErrors()[2])
		assert.Equal(t, "Validation errors: error 1; error 2; error 3", validationErrors.Error())
	})

	t.Run("AddIf with true condition", func(t *testing.T) {
		validationErrors := NewValidationErrors()
		validationErrors.AddIf(true, "conditional error")

		assert.True(t, validationErrors.HasErrors())
		assert.Len(t, validationErrors.GetErrors(), 1)
		assert.Equal(t, "conditional error", validationErrors.GetErrors()[0])
	})

	t.Run("AddIf with false condition", func(t *testing.T) {
		validationErrors := NewValidationErrors()
		validationErrors.AddIf(false, "conditional error")

		assert.False(t, validationErrors.HasErrors())
		assert.Empty(t, validationErrors.GetErrors())
	})

	t.Run("Mix of Add and AddIf", func(t *testing.T) {
		validationErrors := NewValidationErrors()
		validationErrors.Add("direct error")
		validationErrors.AddIf(true, "true conditional error")
		validationErrors.AddIf(false, "false conditional error")

		assert.True(t, validationErrors.HasErrors())
		assert.Len(t, validationErrors.GetErrors(), 2)
		assert.Equal(t, "direct error", validationErrors.GetErrors()[0])
		assert.Equal(t, "true conditional error", validationErrors.GetErrors()[1])
		assert.Equal(t, "Validation errors: direct error; true conditional error", validationErrors.Error())
	})
}
