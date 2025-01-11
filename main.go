package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/goodfoodcesi/auth-api/crypto"
	"github.com/goodfoodcesi/auth-api/domain/service"
	"github.com/goodfoodcesi/auth-api/infrastructure/database"
	"github.com/goodfoodcesi/auth-api/infrastructure/database/repository"
	"github.com/goodfoodcesi/auth-api/infrastructure/jwt"
	"github.com/goodfoodcesi/auth-api/infrastructure/logger"
	"github.com/goodfoodcesi/auth-api/interfaces/http/handler"
	"github.com/goodfoodcesi/auth-api/interfaces/http/router"
	"go.uber.org/zap"
)

func main() {
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

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

	tokenManager := jwt.NewTokenManager(
		os.Getenv("JWT_ACCESS_SECRET"),
		os.Getenv("JWT_REFRESH_SECRET"),
		time.Hour,    // Access token TTL
		time.Hour*24, // Refresh token TTL
	)

	userRepo := repository.NewUserRepository(db)

	// Initialiser les services
	passwordManager := crypto.NewPasswordManager(os.Getenv("PASSWORD_SECRET"))
	userService := service.NewUserService(userRepo, tokenManager, passwordManager)

	// Initialiser les handlers
	userHandler := handler.NewUserHandler(userService, logger)

	// Configurer le routeur
	r := router.NewRouter(userHandler, logger, tokenManager)

	// Démarrer le serveur
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	logger.Info("Server starting on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
