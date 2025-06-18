package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"time"
)

// RequestFilter define os filtros para listar pedidos.
type RequestFilter struct {
	Status      *string
	Destination *string
	StartDate   *time.Time
	EndDate     *time.Time
	UserID      *uuid.UUID // Para filtrar por usuário específico
}

// RequestRepository é a interface (porta) para o repositório de pedidos.
type RequestRepository interface {
	Create(ctx context.Context, request *model.TravelRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.TravelRequest, error)
	FindAll(ctx context.Context, filter RequestFilter) ([]model.TravelRequest, error)
	Update(ctx context.Context, request *model.TravelRequest) error
}
