package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/goodfoodcesi/auth-api/db/sqlc"
	"github.com/goodfoodcesi/auth-api/schema"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "Failed payload", "error": err.Error()})
		return
	}
	now := time.Now()
	fmt.Println(now)

	args := &db.CreateUserParams{
		FirstName:   payload.FirstName,
		LastName:    payload.LastName,
		PhoneNumber: payload.PhoneNumber,
		UserRole:    db.USERROLE(payload.UserRole),
		UpdatedAt:   pgtype.Timestamp{Time: now, Valid: true},
	}

	user, err := uc.db.CreateUser(ctx, *args)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": "Failed to create user", "error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": user})
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
