package ports

import (
	"context"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
)

// TravelService é a interface (porta) para o serviço de viagens.
type TravelService interface {
	Create(ctx context.Context, req *model.TravelRequest, user *model.User) error
	UpdateStatus(ctx context.Context, requestID uuid.UUID, newStatus model.Status, updater *model.User) (*model.TravelRequest, error)
	FindByID(ctx context.Context, requestID uuid.UUID, user *model.User) (*model.TravelRequest, error)
	FindAll(ctx context.Context, filter RequestFilter, user *model.User) ([]model.TravelRequest, error)
	Cancel(ctx context.Context, requestID uuid.UUID, user *model.User) (*model.TravelRequest, error)
}

// UserService é a interface (porta) para o serviço de usuários.
type UserService interface {
	RegisterUser(ctx context.Context, username, password, role string) (*model.User, error)
	AuthenticateUser(ctx context.Context, username, password string) (*model.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error)
}

// Notifier é a interface (porta) para o serviço de notificações.
type Notifier interface {
	SendStatusUpdate(ctx context.Context, request model.TravelRequest) error
}
