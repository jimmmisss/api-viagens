package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"github.com/jimmmmisss/api-viagens/internal/domain/ports"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/out/mysql/db"
	"strings"
)

// SQLCRepository implementa a interface RequestRepository.
type SQLCRepository struct {
	db      *sql.DB
	queries *db.Queries
}

// NewSQLCRepository cria uma nova instância de SQLCRepository.
func NewSQLCRepository(dbConn *sql.DB) *SQLCRepository {
	return &SQLCRepository{
		db:      dbConn,
		queries: db.New(dbConn),
	}
}

// Save salva um pedido de viagem no banco de dados.
func (r *SQLCRepository) Create(ctx context.Context, request *model.TravelRequest) error {
	dbRequest := FromDomainRequest(request)

	// Se o pedido não existe, insere
	err := r.queries.CreateRequest(ctx, db.CreateRequestParams{
		ID:            dbRequest.ID,
		RequesterName: dbRequest.RequesterName,
		Destination:   dbRequest.Destination,
		DepartureDate: dbRequest.DepartureDate,
		ReturnDate:    dbRequest.ReturnDate,
		Status:        dbRequest.Status,
		CreatedAt:     dbRequest.CreatedAt,
		UserID:        dbRequest.UserID,
	})
	if err != nil {
		return fmt.Errorf("erro ao inserir pedido: %w", err)
	}

	return nil
}

// FindByID busca um pedido de viagem pelo ID.
func (r *SQLCRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TravelRequest, error) {
	travelRequest, err := r.queries.GetRequestByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("pedido não encontrado: %w", err)
		}
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	// Converter para o modelo de domínio
	dbRequest := &DBTravelRequest{
		ID:            travelRequest.ID,
		RequesterName: travelRequest.RequesterName,
		Destination:   travelRequest.Destination,
		DepartureDate: travelRequest.DepartureDate,
		ReturnDate:    travelRequest.ReturnDate,
		Status:        travelRequest.Status,
		CreatedAt:     travelRequest.CreatedAt,
		UserID:        travelRequest.UserID,
	}

	return dbRequest.ToDomain()
}

// Update atualiza um pedido de viagem no banco de dados.
func (r *SQLCRepository) Update(ctx context.Context, request *model.TravelRequest) error {
	// Converte para o modelo de banco de dados
	dbRequest := FromDomainRequest(request)

	// Atualiza o status do pedido
	err := r.queries.UpdateRequestStatus(ctx, db.UpdateRequestStatusParams{
		ID:     dbRequest.ID,
		Status: dbRequest.Status,
	})
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do pedido: %w", err)
	}

	return nil
}

// FindAll busca pedidos de viagem com filtros.
func (r *SQLCRepository) FindAll(ctx context.Context, filter ports.RequestFilter) ([]model.TravelRequest, error) {
	var travelRequests []db.TravelRequests
	var err error

	// Se tiver filtro de usuário, usa o método específico
	if filter.UserID != nil && filter.Status == nil && filter.Destination == nil && filter.StartDate == nil && filter.EndDate == nil {
		travelRequests, err = r.queries.ListRequestsByUserID(ctx, filter.UserID.String())
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar pedidos por usuário: %w", err)
		}
	} else {
		// Caso contrário, busca todos e filtra em memória
		travelRequests, err = r.queries.ListAllRequests(ctx)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar todos os pedidos: %w", err)
		}

		// Aplicar filtros em memória
		var filteredRequests []db.TravelRequests
		for _, req := range travelRequests {
			// Filtro de status
			if filter.Status != nil && req.Status != *filter.Status {
				continue
			}

			// Filtro de destino
			if filter.Destination != nil && !strings.Contains(strings.ToLower(req.Destination), strings.ToLower(*filter.Destination)) {
				continue
			}

			// Filtro de data de partida
			if filter.StartDate != nil && req.DepartureDate.Before(*filter.StartDate) {
				continue
			}

			// Filtro de data de retorno
			if filter.EndDate != nil && req.ReturnDate.After(*filter.EndDate) {
				continue
			}

			// Filtro de usuário
			if filter.UserID != nil && req.UserID != filter.UserID.String() {
				continue
			}

			filteredRequests = append(filteredRequests, req)
		}

		travelRequests = filteredRequests
	}

	// Converter para domínio
	var domainRequests []model.TravelRequest
	for _, req := range travelRequests {
		// Converter para o modelo de domínio
		dbRequest := &DBTravelRequest{
			ID:            req.ID,
			RequesterName: req.RequesterName,
			Destination:   req.Destination,
			DepartureDate: req.DepartureDate,
			ReturnDate:    req.ReturnDate,
			Status:        req.Status,
			CreatedAt:     req.CreatedAt,
			UserID:        req.UserID,
		}

		domainReq, err := dbRequest.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("erro ao converter pedido para domínio: %w", err)
		}
		domainRequests = append(domainRequests, *domainReq)
	}

	return domainRequests, nil
}

// FindUserByUsername busca um usuário pelo nome de usuário.
func (r *SQLCRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("usuário não encontrado: %w", err)
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// Converter para o modelo de domínio
	dbUser := &DBUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		CreatedAt:    user.CreatedAt,
	}

	return dbUser.ToDomain()
}

// FindUserByID busca um usuário pelo ID.
func (r *SQLCRepository) FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := r.queries.GetUserByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("usuário não encontrado: %w", err)
		}
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// Converter para o modelo de domínio
	dbUser := &DBUser{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		CreatedAt:    user.CreatedAt,
	}

	return dbUser.ToDomain()
}

// SaveUser salva um usuário no banco de dados.
func (r *SQLCRepository) SaveUser(ctx context.Context, user *model.User) error {
	dbUser := FromDomainUser(user)

	// Verifica se o usuário já existe
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT 1 FROM users WHERE id = ?", dbUser.ID).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("erro ao verificar existência do usuário: %w", err)
	}

	// Se o usuário já existe, atualiza
	if exists {
		// Não há método de atualização de usuário gerado pelo SQLc, então usamos SQL direto
		_, err = r.db.ExecContext(ctx,
			"UPDATE users SET username = ?, password_hash = ?, role = ? WHERE id = ?",
			dbUser.Username, dbUser.PasswordHash, dbUser.Role, dbUser.ID)
		if err != nil {
			return fmt.Errorf("erro ao atualizar usuário: %w", err)
		}
		return nil
	}

	// Se o usuário não existe, insere
	err = r.queries.CreateUser(ctx, db.CreateUserParams{
		ID:           dbUser.ID,
		Username:     dbUser.Username,
		PasswordHash: dbUser.PasswordHash,
		Role:         dbUser.Role,
	})
	if err != nil {
		return fmt.Errorf("erro ao inserir usuário: %w", err)
	}

	return nil
}
