package main

import (
	"context"
	db "github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"os"

	"github.com/goodfoodcesi/auth-api/crypto"
	"github.com/goodfoodcesi/auth-api/infrastructure/database"
	"github.com/goodfoodcesi/auth-api/infrastructure/database/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	dbConfig := database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}

	dbConn, err := database.NewPostgresPool(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConn.Close()

	pwdManager := crypto.NewPasswordManager(os.Getenv("SECRET_KEY"))
	hashedPassword, err := pwdManager.HashPassword("admin123")
	if err != nil {
		logger.Fatal("Failed to hash password", zap.Error(err))
	}

	adminUser := &db.User{
		ID:           pgtype.UUID{Bytes: uuid.New()},
		Firstname:    "Admin",
		Lastname:     "User",
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		Role:         db.UserRoleAdmin,
	}

	userRepo := repository.NewUserRepository(dbConn)
	if _, err := userRepo.Create(context.Background(), adminUser); err != nil {
		log.Fatal(err)
	}

	logger.Info("Admin user created successfully")
}
