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

// setupTripTestDB creates a connection to the test database and sets up the trips table
func setupTripTestDB(t *testing.T) *pgxpool.Pool {
	// Use environment variables or a fixed connection string for tests
	connStr := "postgres://user:password@localhost:5432/tripdb_test?sslmode=disable"

	dbpool, err := pgxpool.New(context.Background(), connStr)
	require.NoError(t, err, "Failed to connect to test database")

	// Create test tables if they don't exist
	_, err = dbpool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS trips (
			id UUID PRIMARY KEY,
			requester_id UUID NOT NULL,
			destination TEXT NOT NULL,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create test table")

	// Clean up existing test data
	_, err = dbpool.Exec(context.Background(), "DELETE FROM trips")
	require.NoError(t, err, "Failed to clean up test data")

	return dbpool
}

func TestPostgresTripRepository_Create(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	dbpool := setupTripTestDB(t)
	defer dbpool.Close()

	repo := repository.NewPostgresTripRepository(dbpool)
	ctx := context.Background()

	// Test data
	tripID := uuid.New()
	requesterID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond) // PostgreSQL truncates to microseconds
	startDate := now.AddDate(0, 1, 0)                  // 1 month from now
	endDate := startDate.AddDate(0, 0, 7)              // 7 days after start

	trip := &domain.Trip{
		ID:          tripID,
		RequesterID: requesterID,
		Destination: "Paris",
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      domain.StatusRequested,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test Create
	err := repo.Create(ctx, trip)
	assert.NoError(t, err)

	// Verify the trip was created
	var count int
	err = dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM trips WHERE id = $1", tripID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPostgresTripRepository_FindByID(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	dbpool := setupTripTestDB(t)
	defer dbpool.Close()

	repo := repository.NewPostgresTripRepository(dbpool)
	ctx := context.Background()

	// Test data
	tripID := uuid.New()
	requesterID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	startDate := now.AddDate(0, 1, 0)
	endDate := startDate.AddDate(0, 0, 7)

	// Insert test trip
	_, err := dbpool.Exec(ctx, `
		INSERT INTO trips (id, requester_id, destination, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, tripID, requesterID, "Paris", startDate, endDate, domain.StatusRequested, now, now)
	require.NoError(t, err)

	// Test FindByID - existing trip
	trip, err := repo.FindByID(ctx, tripID)
	assert.NoError(t, err)
	assert.NotNil(t, trip)
	assert.Equal(t, tripID, trip.ID)
	assert.Equal(t, requesterID, trip.RequesterID)
	assert.Equal(t, "Paris", trip.Destination)
	assert.Equal(t, startDate, trip.StartDate)
	assert.Equal(t, endDate, trip.EndDate)
	assert.Equal(t, domain.StatusRequested, trip.Status)
	assert.Equal(t, now, trip.CreatedAt)
	assert.Equal(t, now, trip.UpdatedAt)

	// Test FindByID - non-existent trip
	nonExistentID := uuid.New()
	trip, err = repo.FindByID(ctx, nonExistentID)
	assert.NoError(t, err) // Not finding a trip is not an error
	assert.Nil(t, trip)
}

func TestPostgresTripRepository_List(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	dbpool := setupTripTestDB(t)
	defer dbpool.Close()

	repo := repository.NewPostgresTripRepository(dbpool)
	ctx := context.Background()

	// Test data - create multiple trips with different properties
	requesterID1 := uuid.New()
	requesterID2 := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Trip 1: Paris, Requested, for requester 1
	trip1ID := uuid.New()
	trip1StartDate := now.AddDate(0, 1, 0)
	trip1EndDate := trip1StartDate.AddDate(0, 0, 7)

	// Trip 2: London, Approved, for requester 1
	trip2ID := uuid.New()
	trip2StartDate := now.AddDate(0, 2, 0)
	trip2EndDate := trip2StartDate.AddDate(0, 0, 10)

	// Trip 3: Rome, Canceled, for requester 2
	trip3ID := uuid.New()
	trip3StartDate := now.AddDate(0, 3, 0)
	trip3EndDate := trip3StartDate.AddDate(0, 0, 5)

	// Insert test trips
	_, err := dbpool.Exec(ctx, `
		INSERT INTO trips (id, requester_id, destination, start_date, end_date, status, created_at, updated_at)
		VALUES 
		($1, $2, $3, $4, $5, $6, $7, $8),
		($9, $10, $11, $12, $13, $14, $15, $16),
		($17, $18, $19, $20, $21, $22, $23, $24)
	`,
		trip1ID, requesterID1, "Paris", trip1StartDate, trip1EndDate, domain.StatusRequested, now, now,
		trip2ID, requesterID1, "London", trip2StartDate, trip2EndDate, domain.StatusApproved, now.Add(time.Hour), now.Add(time.Hour),
		trip3ID, requesterID2, "Rome", trip3StartDate, trip3EndDate, domain.StatusCanceled, now.Add(2*time.Hour), now.Add(2*time.Hour))
	require.NoError(t, err)

	// Test List - all trips
	trips, err := repo.List(ctx, domain.ListTripsParams{})
	assert.NoError(t, err)
	assert.Len(t, trips, 3)

	// Test List - filter by requester ID
	trips, err = repo.List(ctx, domain.ListTripsParams{
		RequesterID: &requesterID1,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 2)
	for _, trip := range trips {
		assert.Equal(t, requesterID1, trip.RequesterID)
	}

	// Test List - filter by status
	approvedStatus := domain.StatusApproved
	trips, err = repo.List(ctx, domain.ListTripsParams{
		Status: &approvedStatus,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, domain.StatusApproved, trips[0].Status)

	// Test List - filter by destination
	londonDest := "London"
	trips, err = repo.List(ctx, domain.ListTripsParams{
		Destination: &londonDest,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, "London", trips[0].Destination)

	// Test List - filter by start date
	trips, err = repo.List(ctx, domain.ListTripsParams{
		StartDate: &trip2StartDate,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 2) // Should return trips 2 and 3 (start date >= trip2StartDate)

	// Test List - filter by end date
	trips, err = repo.List(ctx, domain.ListTripsParams{
		EndDate: &trip1EndDate,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 1) // Should return only trip 1 (end date <= trip1EndDate)

	// Test List - multiple filters
	trips, err = repo.List(ctx, domain.ListTripsParams{
		RequesterID: &requesterID1,
		Status:      &approvedStatus,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 1)
	assert.Equal(t, requesterID1, trips[0].RequesterID)
	assert.Equal(t, domain.StatusApproved, trips[0].Status)

	// Test List - no matching trips
	nonExistentID := uuid.New()
	trips, err = repo.List(ctx, domain.ListTripsParams{
		RequesterID: &nonExistentID,
	})
	assert.NoError(t, err)
	assert.Len(t, trips, 0)
}

func TestPostgresTripRepository_UpdateStatus(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup
	dbpool := setupTripTestDB(t)
	defer dbpool.Close()

	repo := repository.NewPostgresTripRepository(dbpool)
	ctx := context.Background()

	// Test data
	tripID := uuid.New()
	requesterID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	startDate := now.AddDate(0, 1, 0)
	endDate := startDate.AddDate(0, 0, 7)

	// Insert test trip with status "requested"
	_, err := dbpool.Exec(ctx, `
		INSERT INTO trips (id, requester_id, destination, start_date, end_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, tripID, requesterID, "Paris", startDate, endDate, domain.StatusRequested, now, now)
	require.NoError(t, err)

	// Test UpdateStatus - change to approved
	err = repo.UpdateStatus(ctx, tripID, domain.StatusApproved)
	assert.NoError(t, err)

	// Verify the status was updated
	var status string
	var updatedAt time.Time
	err = dbpool.QueryRow(ctx, "SELECT status, updated_at FROM trips WHERE id = $1", tripID).Scan(&status, &updatedAt)
	assert.NoError(t, err)
	assert.Equal(t, string(domain.StatusApproved), status)
	assert.True(t, updatedAt.After(now), "updated_at should be updated")

	// Test UpdateStatus - change to canceled
	err = repo.UpdateStatus(ctx, tripID, domain.StatusCanceled)
	assert.NoError(t, err)

	// Verify the status was updated again
	err = dbpool.QueryRow(ctx, "SELECT status FROM trips WHERE id = $1", tripID).Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, string(domain.StatusCanceled), status)

	// Test UpdateStatus - non-existent trip
	nonExistentID := uuid.New()
	err = repo.UpdateStatus(ctx, nonExistentID, domain.StatusApproved)
	assert.NoError(t, err) // Should not error, but also not update anything

	// Verify no new trips were created
	var count int
	err = dbpool.QueryRow(ctx, "SELECT COUNT(*) FROM trips").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
