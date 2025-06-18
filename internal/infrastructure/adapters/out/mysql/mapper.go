package mysql

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"time"
)

// DBTravelRequest é a representação de um pedido de viagem no banco de dados.
type DBTravelRequest struct {
	ID            string
	RequesterName string
	Destination   string
	DepartureDate time.Time
	ReturnDate    time.Time
	Status        string
	CreatedAt     time.Time
	UserID        string
}

// ToDomain converte um DBTravelRequest para um model.TravelRequest.
func (r *DBTravelRequest) ToDomain() (*model.TravelRequest, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID: %w", err)
	}

	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter UserID: %w", err)
	}

	return &model.TravelRequest{
		ID:            id,
		RequesterName: r.RequesterName,
		Destination:   r.Destination,
		DepartureDate: r.DepartureDate,
		ReturnDate:    r.ReturnDate,
		Status:        model.Status(r.Status),
		CreatedAt:     r.CreatedAt,
		UserID:        userID,
	}, nil
}

// FromDomainRequest converte um model.TravelRequest para um DBTravelRequest.
func FromDomainRequest(r *model.TravelRequest) *DBTravelRequest {
	return &DBTravelRequest{
		ID:            r.ID.String(),
		RequesterName: r.RequesterName,
		Destination:   r.Destination,
		DepartureDate: r.DepartureDate,
		ReturnDate:    r.ReturnDate,
		Status:        string(r.Status),
		CreatedAt:     r.CreatedAt,
		UserID:        r.UserID.String(),
	}
}

// ScanTravelRequest escaneia uma linha do banco de dados para um DBTravelRequest.
func ScanTravelRequest(row *sql.Row) (*DBTravelRequest, error) {
	var req DBTravelRequest
	err := row.Scan(
		&req.ID,
		&req.RequesterName,
		&req.Destination,
		&req.DepartureDate,
		&req.ReturnDate,
		&req.Status,
		&req.CreatedAt,
		&req.UserID,
	)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// ScanTravelRequests escaneia múltiplas linhas do banco de dados para DBTravelRequests.
func ScanTravelRequests(rows *sql.Rows) ([]*DBTravelRequest, error) {
	var requests []*DBTravelRequest
	for rows.Next() {
		var req DBTravelRequest
		err := rows.Scan(
			&req.ID,
			&req.RequesterName,
			&req.Destination,
			&req.DepartureDate,
			&req.ReturnDate,
			&req.Status,
			&req.CreatedAt,
			&req.UserID,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

// DBUser é a representação de um usuário no banco de dados.
type DBUser struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

// ToDomain converte um DBUser para um model.User.
func (u *DBUser) ToDomain() (*model.User, error) {
	id, err := uuid.Parse(u.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter ID: %w", err)
	}

	return &model.User{
		ID:           id,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
	}, nil
}

// FromDomainUser converte um model.User para um DBUser.
func FromDomainUser(u *model.User) *DBUser {
	return &DBUser{
		ID:           u.ID.String(),
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
	}
}

// ScanUser escaneia uma linha do banco de dados para um DBUser.
func ScanUser(row *sql.Row) (*DBUser, error) {
	var user DBUser
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
