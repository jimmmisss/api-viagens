package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"github.com/jimmmmisss/api-viagens/internal/domain/ports"
)

var (
	ErrRequestNotFound        = errors.New("travel request not found")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrCannotUpdateOwn        = errors.New("user cannot approve or cancel their own request")
	ErrCancellationNotAllowed = errors.New("cancellation of this approved request is not allowed")
)

// TravelService implementa os casos de uso.
type TravelService struct {
	repo     ports.RequestRepository
	notifier ports.Notifier
}

func NewTravelService(repo ports.RequestRepository, notifier ports.Notifier) *TravelService {
	return &TravelService{repo: repo, notifier: notifier}
}

// CreateRequest cria um novo pedido de viagem.
func (s *TravelService) Create(ctx context.Context, req *model.TravelRequest, user *model.User) error {
	req.ID = uuid.New()
	req.Status = model.StatusRequested
	req.UserID = user.ID
	req.RequesterName = user.Username
	return s.repo.Create(ctx, req)
}

// UpdateRequestStatus atualiza o status de um pedido.
func (s *TravelService) UpdateStatus(ctx context.Context, requestID uuid.UUID, newStatus model.Status, updater *model.User) (*model.TravelRequest, error) {
	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, ErrRequestNotFound
	}

	if updater.Role != "manager" {
		return nil, ErrPermissionDenied
	}
	if req.UserID == updater.ID {
		return nil, ErrCannotUpdateOwn
	}

	if newStatus == model.StatusApproved {
		req.Approve()
	} else if newStatus == model.StatusCanceled {
		req.Cancel()
	} else {
		return nil, errors.New("invalid status")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	s.notifier.SendStatusUpdate(ctx, *req)
	return req, nil
}

// GetRequestByID obtém um pedido de viagem pelo ID.
func (s *TravelService) FindByID(ctx context.Context, requestID uuid.UUID, user *model.User) (*model.TravelRequest, error) {
	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, ErrRequestNotFound
	}

	// Verifica se o usuário tem permissão para ver este pedido
	// Gerentes podem ver qualquer pedido, usuários comuns só podem ver seus próprios pedidos
	if user.Role != "manager" && req.UserID != user.ID {
		return nil, ErrPermissionDenied
	}

	return req, nil
}

// ListRequests lista pedidos de viagem com filtros.
func (s *TravelService) FindAll(ctx context.Context, filter ports.RequestFilter, user *model.User) ([]model.TravelRequest, error) {
	// Se não for gerente, força o filtro para mostrar apenas os pedidos do próprio usuário
	if user.Role != "manager" {
		filter.UserID = &user.ID
	}

	return s.repo.FindAll(ctx, filter)
}

// CancelRequest cancela um pedido de viagem.
func (s *TravelService) Cancel(ctx context.Context, requestID uuid.UUID, user *model.User) (*model.TravelRequest, error) {
	req, err := s.repo.FindByID(ctx, requestID)
	if err != nil {
		return nil, ErrRequestNotFound
	}

	// Usuários comuns só podem cancelar seus próprios pedidos
	if user.Role != "manager" && req.UserID != user.ID {
		return nil, ErrPermissionDenied
	}

	// Verifica se o pedido pode ser cancelado
	if req.Status == model.StatusApproved && !req.CanBeCanceled() {
		return nil, ErrCancellationNotAllowed
	}

	if err := req.Cancel(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	s.notifier.SendStatusUpdate(ctx, *req)
	return req, nil
}
