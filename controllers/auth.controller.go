package controllers

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/goodfoodcesi/auth-api/db/sqlc"
	"github.com/goodfoodcesi/auth-api/schema"
	"github.com/goodfoodcesi/auth-api/utils"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	db        *db.Queries
	ctx       context.Context
	jwtSecret string
}

func NewAuthController(db *db.Queries, ctx context.Context, jwtSecret string) *AuthController {
	return &AuthController{db, ctx, jwtSecret}
}

func (ac *AuthController) Login(ctx *gin.Context) {
	var payload *schema.LoginRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "validation_error",
			"errors": getValidationErrors(err),
		})
		return
	}

	// Normalisation de l'email
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))

	// Recherche de l'utilisateur
	user, err := ac.db.GetUserByEmailWithPassword(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "Invalid credentials",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process login",
		})
		return
	}

	// Vérification du mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid credentials",
		})
		return
	}

	userID := hex.EncodeToString(user.UserID.Bytes[:])
	// Génération des tokens
	accessToken, err := utils.GenerateToken(userID, string(user.UserRole), ac.jwtSecret, time.Hour*24)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to generate access token",
		})
		return
	}

	refreshToken, err := utils.GenerateToken(userID, string(user.UserRole), ac.jwtSecret, time.Hour*24*7)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to generate refresh token",
		})
		return
	}

	response := schema.LoginResponse{
		User: schema.UserResponse{
			UserID:      userID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			UserRole:    string(user.UserRole),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}
