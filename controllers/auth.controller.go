package controllers

import (
	"context"
	"errors"
	"fmt"
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
	db            *db.Queries
	ctx           context.Context
	jwtSecret     string
	refreshSecret string
}

func NewAuthController(db *db.Queries, ctx context.Context, jwtSecret string, refreshSecret string) *AuthController {
	return &AuthController{db, ctx, jwtSecret, refreshSecret}
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

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"error":  "Invalid credentials",
		})
		return
	}

	userID := fmt.Sprintf("%x-%x-%x-%x-%x", user.UserID.Bytes[0:4], user.UserID.Bytes[4:6], user.UserID.Bytes[6:8], user.UserID.Bytes[8:10], user.UserID.Bytes[10:16])
	accessToken, err := utils.GenerateToken(userID, string(user.UserRole), ac.jwtSecret, time.Hour*24)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to generate access token",
		})
		return
	}

	refreshToken, err := utils.GenerateToken(userID, string(user.UserRole), ac.refreshSecret, time.Hour*24*7)
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
