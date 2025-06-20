package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jimmmmisss/api-viagens/internal/config"
	"github.com/jimmmmisss/api-viagens/internal/domain"
	"github.com/jimmmmisss/api-viagens/internal/service"
	"github.com/jimmmmisss/api-viagens/internal/utils"
)

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) RegisterUser(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := parseValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors.GetErrors()})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		// Check if it's a validation error
		if validationErrs, ok := err.(*domain.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) LoginUser(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := parseValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors.GetErrors()})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Check if it's a validation error
		if validationErrs, ok := err.(*domain.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrs.GetErrors()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	cfg, _ := config.Load() // In a real app, inject config or get from context
	token, err := utils.GenerateJWT(user.ID, cfg.JWTSecretKey, cfg.JWTExpirationHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) LogoutUser(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header format must be Bearer {token}"})
		return
	}

	tokenString := parts[1]

	// Parse the token to get the expiration time
	token, err := utils.ValidateJWT(tokenString, h.jwtSecretKey)
	if err != nil {
		// Even if the token is invalid, we'll add it to the blacklist
		// This is a defensive measure
		utils.GetTokenBlacklist().Add(tokenString, time.Now().Add(24*time.Hour))
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
		return
	}

	// Get the expiration time from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.GetTokenBlacklist().Add(tokenString, time.Now().Add(24*time.Hour))
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
		return
	}

	// Get the expiration time from the claims
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		utils.GetTokenBlacklist().Add(tokenString, time.Now().Add(24*time.Hour))
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
		return
	}

	// Add the token to the blacklist with its expiration time
	expTime := time.Unix(int64(expFloat), 0)
	utils.GetTokenBlacklist().Add(tokenString, expTime)

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetCurrentUser returns the currently logged in user's data
func (h *Handler) GetCurrentUser(c *gin.Context) {
	// Get the user ID from the context (set by the auth middleware)
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the user from the database
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user data"})
		return
	}

	// Return the user data
	c.JSON(http.StatusOK, user)
}
