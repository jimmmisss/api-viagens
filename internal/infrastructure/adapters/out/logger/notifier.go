package logger

import (
	"context"
	"log"

	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"github.com/jimmmmisss/api-viagens/internal/domain/ports"
)

// ConsoleNotifier implementa a interface Notifier para enviar notificações via console.
type ConsoleNotifier struct{}

// NewConsoleNotifier cria uma nova instância de ConsoleNotifier.
func NewConsoleNotifier() ports.Notifier {
	return &ConsoleNotifier{}
}

// SendStatusUpdate envia uma notificação de atualização de status.
func (n *ConsoleNotifier) SendStatusUpdate(ctx context.Context, request model.TravelRequest) error {
	log.Printf("[NOTIFICAÇÃO] Pedido %s atualizado para status: %s", request.ID, request.Status)
	return nil
}
