package domain

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Don't expose password hash
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Validate checks if the user data is valid according to business rules
func (u *User) Validate() error {
	validationErrors := NewValidationErrors()

	// Check required fields
	validationErrors.AddIf(u.Name == "", "name is required")
	validationErrors.AddIf(u.Email == "", "email is required")

	// Only check email format if email is not empty
	if u.Email != "" {
		// Trim spaces from email and check if it's still valid
		validationErrors.AddIf(strings.TrimSpace(u.Email) != u.Email, "email cannot contain leading or trailing spaces")

		// Validate email format
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		validationErrors.AddIf(!emailRegex.MatchString(u.Email), "invalid email format")
	}

	validationErrors.AddIf(u.PasswordHash == "", "password_hash is required")

	if validationErrors.HasErrors() {
		return validationErrors
	}
	return nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
}
