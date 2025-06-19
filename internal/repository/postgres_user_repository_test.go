package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a connection to the test database
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Use environment variables or a fixed connection string for tests
	// In a real project, you might want to use a separate test database or schema
	connStr := "postgres://user:password@localhost:5432/tripdb_test?sslmode=disable"
	
	dbpool, err := pgxpool.New(context.Background(), connStr)
	require.NoError(t, err, "Failed to connect to test database")
	
	// Create test tables if they don't exist
	_, err = dbpool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create test table")
	
	// Clean up existing test data
	_, err = dbpool.Exec(context.Background(), "DELETE FROM users")
	require.NoError(t, err, "Failed to clean up test data")
	
	return dbpool
}

func TestPostgresUserRepository_Create(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	// Setup
	dbpool := setupTestDB(t)
	defer dbpool.Close()
	
	repo := repository.NewPostgresUserRepository(dbpool)
	ctx := context.Background()
	
	// Test data
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond) // PostgreSQL truncates to microseconds
	user := &domain.User{
		ID:           userID,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	// Test Create
	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	
	// Verify the user was created
	var count int
	err = dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	
	// Test duplicate email
	duplicateUser := &domain.User{
		ID:           uuid.New(),
		Name:         "Another User",
		Email:        "test@example.com", // Same email
		PasswordHash: "another_hashed_password",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	
	err = repo.Create(ctx, duplicateUser)
	assert.Error(t, err) // Should fail due to unique constraint on email
}

func TestPostgresUserRepository_FindByEmail(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	// Setup
	dbpool := setupTestDB(t)
	defer dbpool.Close()
	
	repo := repository.NewPostgresUserRepository(dbpool)
	ctx := context.Background()
	
	// Test data
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := "find-by-email@example.com"
	
	// Insert test user
	_, err := dbpool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "Find By Email User", email, "hashed_password", now, now)
	require.NoError(t, err)
	
	// Test FindByEmail - existing user
	user, err := repo.FindByEmail(ctx, email)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "Find By Email User", user.Name)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
	
	// Test FindByEmail - non-existent user
	user, err = repo.FindByEmail(ctx, "nonexistent@example.com")
	assert.NoError(t, err) // Not finding a user is not an error
	assert.Nil(t, user)
}

func TestPostgresUserRepository_FindByID(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	// Setup
	dbpool := setupTestDB(t)
	defer dbpool.Close()
	
	repo := repository.NewPostgresUserRepository(dbpool)
	ctx := context.Background()
	
	// Test data
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	
	// Insert test user
	_, err := dbpool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "Find By ID User", "find-by-id@example.com", "hashed_password", now, now)
	require.NoError(t, err)
	
	// Test FindByID - existing user
	user, err := repo.FindByID(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "Find By ID User", user.Name)
	assert.Equal(t, "find-by-id@example.com", user.Email)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
	
	// Test FindByID - non-existent user
	nonExistentID := uuid.New()
	user, err = repo.FindByID(ctx, nonExistentID)
	assert.NoError(t, err) // Not finding a user is not an error
	assert.Nil(t, user)
}