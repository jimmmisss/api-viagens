package http

import (
	"net/http"

	"github.com/jimmmmisss/api-viagens/internal/application/service"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/out/auth"
)

// SetupRouter configura as rotas da API.
func SetupRouter(travelService *service.TravelService, authService *auth.JWTAuthService, userRepo auth.UserRepository) http.Handler {
	mux := http.NewServeMux()

	// Cria os handlers
	travelHandler := NewTravelHandler(travelService)

	// Cria o middleware de autenticação
	authMiddleware := NewAuthMiddleware(authService)

	// Rotas públicas
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Implementação simplificada do login
		w.Write([]byte("Endpoint de login"))
	})

	mux.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		// Implementação simplificada do registro
		w.Write([]byte("Endpoint de registro"))
	})

	// Rotas protegidas
	mux.Handle("/api/requests", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			travelHandler.CreateRequestHandler(w, r)
		case http.MethodGet:
			travelHandler.ListRequestsHandler(w, r)
		default:
			http.Error(w, `{"error": "método não permitido"}`, http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/requests/", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "método não permitido"}`, http.StatusMethodNotAllowed)
			return
		}
		travelHandler.GetRequestHandler(w, r)
	})))

	// Rota para cancelar pedidos (qualquer usuário autenticado pode cancelar seus próprios pedidos)
	mux.Handle("/api/requests/cancel", authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "método não permitido"}`, http.StatusMethodNotAllowed)
			return
		}
		travelHandler.CancelRequestHandler(w, r)
	})))

	// Rotas que exigem papel de gerente
	mux.Handle("/api/manager/requests/status", authMiddleware.Authenticate(
		authMiddleware.RequireRole("manager", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				http.Error(w, `{"error": "método não permitido"}`, http.StatusMethodNotAllowed)
				return
			}
			travelHandler.UpdateRequestStatusHandler(w, r)
		})),
	))

	// Middleware para CORS
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	})
}
