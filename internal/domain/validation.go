package domain

import (
	"fmt"
	"strings"
)

// ValidationErrors is a structure to accumulate validation errors
type ValidationErrors struct {
	errors []string
}

// NewValidationErrors creates a new ValidationErrors instance
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		errors: []string{},
	}
}

// Add adds a new validation error
func (v *ValidationErrors) Add(err string) {
	v.errors = append(v.errors, err)
}

// AddIf adds a new validation error if the condition is true
func (v *ValidationErrors) AddIf(condition bool, err string) {
	if condition {
		v.Add(err)
	}
}

// HasErrors returns true if there are any validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns all validation errors as a slice of strings
func (v *ValidationErrors) GetErrors() []string {
	return v.errors
}

// Error implements the error interface and returns all errors as a single string
func (v *ValidationErrors) Error() string {
	if !v.HasErrors() {
		return ""
	}
	return fmt.Sprintf("Validation errors: %s", strings.Join(v.errors, "; "))
}
