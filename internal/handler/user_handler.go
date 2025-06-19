package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
