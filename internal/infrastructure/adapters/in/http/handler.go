package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/application/service"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"github.com/jimmmmisss/api-viagens/internal/domain/ports"
)

// TravelHandler contém a lógica para manipular requisições HTTP.
type TravelHandler struct {
	service *service.TravelService
}

func NewTravelHandler(s *service.TravelService) *TravelHandler {
	return &TravelHandler{service: s}
}

// CreateRequestHandler lida com a criação de pedidos.
func (h *TravelHandler) CreateRequestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair usuário do contexto (adicionado pelo middleware JWT)
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		http.Error(w, `{"error": "usuário não autenticado"}`, http.StatusUnauthorized)
		return
	}

	// 2. Decodificar e validar o corpo da requisição
	var requestDTO struct {
		Destination   string `json:"destination"`
		DepartureDate string `json:"departure_date"`
		ReturnDate    string `json:"return_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestDTO); err != nil {
		http.Error(w, `{"error": "corpo da requisição inválido"}`, http.StatusBadRequest)
		return
	}

	// Validar campos obrigatórios
	if requestDTO.Destination == "" || requestDTO.DepartureDate == "" || requestDTO.ReturnDate == "" {
		http.Error(w, `{"error": "todos os campos são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	// Converter strings de data para time.Time
	departureDate, err := time.Parse("2006-01-02", requestDTO.DepartureDate)
	if err != nil {
		http.Error(w, `{"error": "formato de data de partida inválido"}`, http.StatusBadRequest)
		return
	}

	returnDate, err := time.Parse("2006-01-02", requestDTO.ReturnDate)
	if err != nil {
		http.Error(w, `{"error": "formato de data de retorno inválido"}`, http.StatusBadRequest)
		return
	}

	// Validar lógica de negócio
	if departureDate.Before(time.Now()) {
		http.Error(w, `{"error": "data de partida deve ser futura"}`, http.StatusBadRequest)
		return
	}

	if returnDate.Before(departureDate) {
		http.Error(w, `{"error": "data de retorno deve ser após a data de partida"}`, http.StatusBadRequest)
		return
	}

	// Criar objeto de domínio
	domainReq := &model.TravelRequest{
		Destination:   requestDTO.Destination,
		DepartureDate: departureDate,
		ReturnDate:    returnDate,
		CreatedAt:     time.Now(),
	}

	// 3. Chamar o serviço da aplicação
	err = h.service.Create(r.Context(), domainReq, user)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 4. Retornar a resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      domainReq.ID.String(),
		"message": "Pedido de viagem criado com sucesso",
	})
}

// GetRequestHandler lida com a obtenção de um pedido pelo ID.
func (h *TravelHandler) GetRequestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair usuário do contexto
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		http.Error(w, `{"error": "usuário não autenticado"}`, http.StatusUnauthorized)
		return
	}

	// 2. Extrair ID do pedido da URL
	requestIDStr := r.URL.Query().Get("id")
	if requestIDStr == "" {
		http.Error(w, `{"error": "ID do pedido é obrigatório"}`, http.StatusBadRequest)
		return
	}

	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		http.Error(w, `{"error": "ID do pedido inválido"}`, http.StatusBadRequest)
		return
	}

	// 3. Chamar o serviço da aplicação
	request, err := h.service.FindByID(r.Context(), requestID, user)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrRequestNotFound {
			status = http.StatusNotFound
		} else if err == service.ErrPermissionDenied {
			status = http.StatusForbidden
		}
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), status)
		return
	}

	// 4. Retornar a resposta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

// ListRequestsHandler lida com a listagem de pedidos.
func (h *TravelHandler) ListRequestsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair usuário do contexto
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		http.Error(w, `{"error": "usuário não autenticado"}`, http.StatusUnauthorized)
		return
	}

	// 2. Extrair filtros da query string
	query := r.URL.Query()
	var filter ports.RequestFilter

	if status := query.Get("status"); status != "" {
		filter.Status = &status
	}

	if destination := query.Get("destination"); destination != "" {
		filter.Destination = &destination
	}

	if startDateStr := query.Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := query.Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			filter.EndDate = &endDate
		}
	}

	// 3. Chamar o serviço da aplicação
	requests, err := h.service.FindAll(r.Context(), filter, user)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// 4. Retornar a resposta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// UpdateRequestStatusHandler lida com a atualização do status de um pedido.
func (h *TravelHandler) UpdateRequestStatusHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair usuário do contexto
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		http.Error(w, `{"error": "usuário não autenticado"}`, http.StatusUnauthorized)
		return
	}

	// 2. Decodificar e validar o corpo da requisição
	var updateDTO struct {
		RequestID string `json:"request_id"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		http.Error(w, `{"error": "corpo da requisição inválido"}`, http.StatusBadRequest)
		return
	}

	// Validar campos obrigatórios
	if updateDTO.RequestID == "" || updateDTO.Status == "" {
		http.Error(w, `{"error": "ID do pedido e status são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	requestID, err := uuid.Parse(updateDTO.RequestID)
	if err != nil {
		http.Error(w, `{"error": "ID do pedido inválido"}`, http.StatusBadRequest)
		return
	}

	// Validar status
	var status model.Status
	switch updateDTO.Status {
	case "approved":
		status = model.StatusApproved
	case "canceled":
		status = model.StatusCanceled
	default:
		http.Error(w, `{"error": "status inválido"}`, http.StatusBadRequest)
		return
	}

	// 3. Chamar o serviço da aplicação
	request, err := h.service.UpdateStatus(r.Context(), requestID, status, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrRequestNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrPermissionDenied || err == service.ErrCannotUpdateOwn {
			statusCode = http.StatusForbidden
		}
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), statusCode)
		return
	}

	// 4. Retornar a resposta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

// CancelRequestHandler lida com o cancelamento de um pedido.
func (h *TravelHandler) CancelRequestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair usuário do contexto
	user, ok := r.Context().Value("user").(*model.User)
	if !ok {
		http.Error(w, `{"error": "usuário não autenticado"}`, http.StatusUnauthorized)
		return
	}

	// 2. Extrair ID do pedido da URL
	requestIDStr := r.URL.Query().Get("id")
	if requestIDStr == "" {
		http.Error(w, `{"error": "ID do pedido é obrigatório"}`, http.StatusBadRequest)
		return
	}

	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		http.Error(w, `{"error": "ID do pedido inválido"}`, http.StatusBadRequest)
		return
	}

	// 3. Chamar o serviço da aplicação
	request, err := h.service.Cancel(r.Context(), requestID, user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == service.ErrRequestNotFound {
			statusCode = http.StatusNotFound
		} else if err == service.ErrPermissionDenied {
			statusCode = http.StatusForbidden
		} else if err == service.ErrCancellationNotAllowed {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), statusCode)
		return
	}

	// 4. Retornar a resposta
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}
