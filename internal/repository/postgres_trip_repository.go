package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jimmmmisss/api-viagens/internal/domain"
)

type postgresTripRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTripRepository(db *pgxpool.Pool) domain.TripRepository {
	return &postgresTripRepository{db: db}
}

func (r *postgresTripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	query := `INSERT INTO trips (id, requester_id, destination, start_date, end_date, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, trip.ID, trip.RequesterID, trip.Destination, trip.StartDate, trip.EndDate, trip.Status, trip.CreatedAt, trip.UpdatedAt)
	return err
}

func (r *postgresTripRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Trip, error) {
	query := `SELECT id, requester_id, destination, start_date, end_date, status, created_at, updated_at
			  FROM trips WHERE id = $1`
	var trip domain.Trip
	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.RequesterID, &trip.Destination, &trip.StartDate,
		&trip.EndDate, &trip.Status, &trip.CreatedAt, &trip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &trip, nil
}

func (r *postgresTripRepository) List(ctx context.Context, params domain.ListTripsParams) ([]*domain.Trip, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`SELECT id, requester_id, destination, start_date, end_date, status, created_at, updated_at
							   FROM trips WHERE 1=1`)

	args := []interface{}{}
	argID := 1

	if params.RequesterID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND requester_id = $%d", argID))
		args = append(args, *params.RequesterID)
		argID++
	}
	if params.Status != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argID))
		args = append(args, *params.Status)
		argID++
	}
	if params.Destination != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND destination ILIKE $%d", argID))
		args = append(args, "%"+*params.Destination+"%")
		argID++
	}
	if params.StartDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND start_date >= $%d", argID))
		args = append(args, *params.StartDate)
		argID++
	}
	if params.EndDate != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND end_date <= $%d", argID))
		args = append(args, *params.EndDate)
		argID++
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC")

	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*domain.Trip
	for rows.Next() {
		var trip domain.Trip
		if err := rows.Scan(
			&trip.ID, &trip.RequesterID, &trip.Destination, &trip.StartDate,
			&trip.EndDate, &trip.Status, &trip.CreatedAt, &trip.UpdatedAt,
		); err != nil {
			return nil, err
		}
		trips = append(trips, &trip)
	}

	return trips, nil
}

func (r *postgresTripRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TripStatus) error {
	query := `UPDATE trips SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	return err
}
