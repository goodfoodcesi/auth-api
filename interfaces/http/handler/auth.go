package handler

import (
	"encoding/json"
	"github.com/goodfoodcesi/auth-api/domain/service"
	"github.com/goodfoodcesi/auth-api/interfaces/http/response"
	"go.uber.org/zap"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	tokens, err := h.authService.RefreshToken(r.Context(), input.RefreshToken)
	if err != nil {
		h.logger.Error("failed to refresh token", zap.Error(err))
		response.Error(w, http.StatusUnauthorized, "Invalid refresh token", nil)
		return
	}

	response.JSON(w, http.StatusOK, tokens)
}
