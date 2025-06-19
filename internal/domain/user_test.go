package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_Validate(t *testing.T) {
	// Setup valid user for reuse
	validID := uuid.New()
	validName := "John Doe"
	validEmail := "john.doe@example.com"
	validPasswordHash := "hashed_password_123"
	now := time.Now()

	t.Run("Valid user", func(t *testing.T) {
		user := &User{
			ID:           validID,
			Name:         validName,
			Email:        validEmail,
			PasswordHash: validPasswordHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := user.Validate()
		assert.NoError(t, err)
	})

	t.Run("Missing name", func(t *testing.T) {
		user := &User{
			ID:           validID,
			Name:         "", // Invalid: empty name
			Email:        validEmail,
			PasswordHash: validPasswordHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := user.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("Missing email", func(t *testing.T) {
		user := &User{
			ID:           validID,
			Name:         validName,
			Email:        "", // Invalid: empty email
			PasswordHash: validPasswordHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := user.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("Invalid email format", func(t *testing.T) {
		user := &User{
			ID:           validID,
			Name:         validName,
			Email:        "invalid-email", // Invalid: not a valid email format
			PasswordHash: validPasswordHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := user.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("Email with spaces", func(t *testing.T) {
		user := &User{
			ID:           validID,
			Name:         validName,
			Email:        " john.doe@example.com ", // Invalid: has leading/trailing spaces
			PasswordHash: validPasswordHash,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := user.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email cannot contain leading or trailing spaces")
	})

	t.Run("Missing password hash", func(t *testing.T) {
		user := &User{
			ID:           validID,
			Name:         validName,
			Email:        validEmail,
			PasswordHash: "", // Invalid: empty password hash
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := user.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password_hash is required")
	})
}