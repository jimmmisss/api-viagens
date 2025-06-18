package model

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

// Status define os possíveis estados de um pedido.
type Status string

const (
	StatusRequested Status = "requested"
	StatusApproved  Status = "approved"
	StatusCanceled  Status = "canceled"
)

// TravelRequest é a entidade de negócio principal. Pura, sem tags de DB ou JSON.
type TravelRequest struct {
	ID            uuid.UUID
	RequesterName string
	Destination   string
	DepartureDate time.Time
	ReturnDate    time.Time
	Status        Status
	CreatedAt     time.Time
	UserID        uuid.UUID // ID do usuário que criou
}

// CanBeCanceled define a regra de negócio para cancelar um pedido aprovado.
// Regra: Não pode ser cancelado se a data de partida for em 7 dias ou menos.
func (tr *TravelRequest) CanBeCanceled() bool {
	if tr.Status != StatusApproved {
		return false // Só pode cancelar o que está aprovado
	}
	daysUntilDeparture := time.Until(tr.DepartureDate).Hours() / 24
	return daysUntilDeparture > 7
}

// Approve muda o status para aprovado.
func (tr *TravelRequest) Approve() {
	tr.Status = StatusApproved
}

// Cancel muda o status para cancelado.
func (tr *TravelRequest) Cancel() error {
	if tr.Status == StatusCanceled {
		return errors.New("request is already canceled")
	}
	tr.Status = StatusCanceled
	return nil
}
