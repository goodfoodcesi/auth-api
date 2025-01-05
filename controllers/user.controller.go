package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	db "github.com/goodfoodcesi/auth-api/db/sqlc"
	"github.com/goodfoodcesi/auth-api/schema"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	db  *db.Queries
	ctx context.Context
}

func NewUserController(db *db.Queries, ctx context.Context) *UserController {
	return &UserController{db, ctx}
}

func (uc *UserController) CreateUser(ctx *gin.Context) {
	var payload *schema.CreateUser

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "validation_error",
			"errors": getValidationErrors(err),
		})
		return
	}

	// Normalisation des données
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	payload.FirstName = strings.TrimSpace(payload.FirstName)
	payload.LastName = strings.TrimSpace(payload.LastName)
	payload.PhoneNumber = strings.TrimSpace(payload.PhoneNumber)

	// Vérification de l'email
	existingUserEmail, err := uc.db.GetUserByEmail(ctx, payload.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to check email uniqueness",
		})
		return
	}
	if existingUserEmail.Email != "" {
		ctx.JSON(http.StatusConflict, gin.H{
			"status": "error",
			"error":  "Email already in use",
		})
		return
	}

	existingUserPhoneNumber, err := uc.db.GetUserByPhoneNumber(ctx, payload.PhoneNumber)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to check phone number uniqueness",
		})
		return
	}
	if existingUserPhoneNumber.PhoneNumber != "" {
		ctx.JSON(http.StatusConflict, gin.H{
			"status": "error",
			"error":  "Phone number already in use",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to process password",
		})
		return
	}

	args := &db.CreateUserParams{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,

		Email:       payload.Email,
		PhoneNumber: payload.PhoneNumber,
		Password:    string(hashedPassword),
		UserRole:    "user",
	}

	user, err := uc.db.CreateUser(ctx, *args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create user",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   user,
	})
}

// Fonction utilitaire pour formater les erreurs de validation
func getValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errors[e.Field()] = "This field is required"
			case "email":
				errors[e.Field()] = "Invalid email format"
			case "min":
				errors[e.Field()] = fmt.Sprintf("Minimum length is %s", e.Param())
			case "max":
				errors[e.Field()] = fmt.Sprintf("Maximum length is %s", e.Param())
			case "e164":
				errors[e.Field()] = "Invalid phone number format (must be E.164 format, ex: +33612345678)"
			default:
				errors[e.Field()] = fmt.Sprintf("Failed validation on %s", e.Tag())
			}
		}
		return errors
	}

	return map[string]string{"general": err.Error()}
}

func (uc *UserController) UpdateUser(ctx *gin.Context) {
	var payload *schema.UpdateUser
	userId := ctx.Param("userId")

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed payload", "error": err.Error()})
		return
	}

	args := &db.UpdateUserParams{
		UserID:      pgtype.UUID{Bytes: uuid.MustParse(userId)},
		FirstName:   pgtype.Text{String: payload.FirstName},
		LastName:    pgtype.Text{String: payload.LastName},
		PhoneNumber: pgtype.Text{String: payload.PhoneNumber},
	}

	user, err := uc.db.UpdateUser(ctx, *args)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "Failed to retrieve user with this ID"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "Failed retrieving user", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) GetAllUsers(ctx *gin.Context) {
	users, err := uc.db.ListUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": users})
}

func (uc *UserController) GetUserById(ctx *gin.Context) {
	userId := ctx.Param("userId")

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Invalid user ID format"})
		return
	}

	user, err := uc.db.GetUserById(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (uc *UserController) DeleteUserById(ctx *gin.Context) {
	userId := ctx.Param("userId")

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "failed", "error": "Invalid user ID format"})
		return
	}

	err = uc.db.DeleteUser(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"status": "failed", "message": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "User deleted successfully"})
}
