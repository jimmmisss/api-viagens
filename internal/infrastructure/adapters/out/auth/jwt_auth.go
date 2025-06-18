package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jimmmmisss/api-viagens/internal/domain/model"
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
)

// JWTAuthService implementa a autenticação JWT.
type JWTAuthService struct {
	secretKey  string
	expiration time.Duration
	userRepo   UserRepository
}

// UserRepository é uma interface para acessar usuários.
type UserRepository interface {
	FindUserByUsername(ctx context.Context, username string) (*model.User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	SaveUser(ctx context.Context, user *model.User) error
}

// NewJWTAuthService cria uma nova instância de JWTAuthService.
func NewJWTAuthService(secretKey string, expirationHours int, userRepo UserRepository) *JWTAuthService {
	return &JWTAuthService{
		secretKey:  secretKey,
		expiration: time.Duration(expirationHours) * time.Hour,
		userRepo:   userRepo,
	}
}

// Claims é a estrutura de dados para o token JWT.
type Claims struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

// GenerateToken gera um token JWT para um usuário.
func (s *JWTAuthService) GenerateToken(user *model.User) (string, error) {
	expirationTime := time.Now().Add(s.expiration)
	claims := &Claims{
		UserID:    user.ID.String(),
		Role:      user.Role,
		ExpiresAt: expirationTime,
		IssuedAt:  time.Now(),
	}

	// Serializa os claims para JSON
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar claims: %w", err)
	}

	// Codifica o payload em base64
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Cria um token simples (header.payload.signature)
	// Em um sistema real, você deve usar uma biblioteca JWT adequada
	token := fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.%s.signature", payload)

	return token, nil
}

// ValidateToken valida um token JWT.
func (s *JWTAuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Divide o token em partes
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Decodifica o payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar payload: %w", err)
	}

	// Desserializa os claims
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("erro ao desserializar claims: %w", err)
	}

	// Verifica se o token expirou
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

// GetUserFromToken obtém o usuário a partir de um token JWT.
func (s *JWTAuthService) GetUserFromToken(ctx context.Context, tokenString string) (*model.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("ID de usuário inválido no token: %w", err)
	}

	user, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return user, nil
}

// AuthenticateUser autentica um usuário e gera um token JWT.
func (s *JWTAuthService) AuthenticateUser(ctx context.Context, username, password string) (string, *model.User, error) {
	user, err := s.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// Em um sistema real, você deve verificar a senha com bcrypt ou similar
	// Por simplicidade, estamos apenas comparando as strings
	if user.PasswordHash != password {
		return "", nil, errors.New("senha incorreta")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao gerar token: %w", err)
	}

	return token, user, nil
}
