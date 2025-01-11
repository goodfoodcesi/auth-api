package handler

import (
	"encoding/json"
	"github.com/goodfoodcesi/auth-api/interfaces/http/response"
	"net/http"

	"github.com/goodfoodcesi/auth-api/domain/service"
	"github.com/goodfoodcesi/auth-api/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
	validator   *validator.Validator
	logger      *zap.Logger
}

func NewUserHandler(userService *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.NewValidator(),
		logger:      logger,
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input service.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if validationErrors := h.validator.Validate(input); validationErrors != nil {
		h.logger.Error("validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": validationErrors,
		})
		return
	}

	id, err := uuid.Parse(r.Context().Value("userID").(string))
	if err != nil {
		h.logger.Error("invalid user ID format", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	user, err := h.userService.Update(r.Context(), pgtype.UUID{Bytes: id, Valid: true}, input)
	if err != nil {
		h.logger.Error("failed to update user", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if validationErrors := h.validator.Validate(input); validationErrors != nil {
		h.logger.Error("validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": validationErrors,
		})
		return
	}

	user, err := h.userService.Register(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to register user", zap.Error(err))

		if err.Error() == "email already exists" {
			response.Error(w, http.StatusConflict, "Email already exists", nil)
			return
		}

		response.Error(w, http.StatusInternalServerError, "Failed to register user", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Failed to encode response", nil)
		return
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	tokens, err := h.userService.Login(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to login", zap.Error(err))
		response.Error(w, http.StatusUnauthorized, "Invalid credentials", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tokens)
	if err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Failed to encode response", nil)
		return
	}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	id, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("invalid user ID format", zap.Error(err))
		response.Error(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	user, err := h.userService.GetByID(r.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		h.logger.Error("failed to get user profile", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Failed to get user profile", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
		response.Error(w, http.StatusInternalServerError, "Failed to encode response", nil)
		return
	}
}
