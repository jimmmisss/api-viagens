package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/jimmmmisss/api-viagens/internal/domain/model"
	"github.com/jimmmmisss/api-viagens/internal/infrastructure/adapters/out/auth"
)

// AuthMiddleware é um middleware para autenticação JWT.
type AuthMiddleware struct {
	authService *auth.JWTAuthService
}

// NewAuthMiddleware cria uma nova instância de AuthMiddleware.
func NewAuthMiddleware(authService *auth.JWTAuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Authenticate é um middleware que verifica se o usuário está autenticado.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrai o token do cabeçalho Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "token de autenticação não fornecido"}`, http.StatusUnauthorized)
			return
		}

		// O token deve estar no formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "formato de token inválido"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Valida o token e obtém o usuário
		user, err := m.authService.GetUserFromToken(r.Context(), tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			if err == auth.ErrExpiredToken {
				status = http.StatusUnauthorized
			}
			http.Error(w, `{"error": "token inválido ou expirado"}`, status)
			return
		}

		// Adiciona o usuário ao contexto da requisição
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole é um middleware que verifica se o usuário tem um papel específico.
func (m *AuthMiddleware) RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtém o usuário do contexto
		user, ok := r.Context().Value("user").(*model.User)
		if !ok {
			http.Error(w, `{"error": "usuário não autenticado"}`, http.StatusUnauthorized)
			return
		}

		// Verifica se o usuário tem o papel necessário
		if user.Role != role {
			http.Error(w, `{"error": "permissão negada"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
