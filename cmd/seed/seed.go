package main

import (
	"context"
	"log"
	"os"

	"github.com/goodfoodcesi/auth-api/crypto"
	"github.com/goodfoodcesi/auth-api/domain/entity"
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

	db, err := database.NewPostgresPool(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	pwdManager := crypto.NewPasswordManager(os.Getenv("SECRET_KEY"))
	hashedPassword, err := pwdManager.HashPassword("admin123")
	if err != nil {
		logger.Fatal("Failed to hash password", zap.Error(err))
	}

	adminUser := &entity.User{
		ID:           uuid.New(),
		FirstName:    "Admin",
		LastName:     "User",
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		Role:         entity.RoleAdmin,
	}

	userRepo := repository.NewUserRepository(db)
	if err := userRepo.Create(context.Background(), adminUser); err != nil {
		log.Fatal(err)
	}

	logger.Info("Admin user created successfully")
}
