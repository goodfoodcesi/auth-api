package main

import (
	"context"
	"github.com/goodfoodcesi/auth-api/infrastructure/database/repository"
	"github.com/goodfoodcesi/auth-api/infrastructure/messaging/rabbitmq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goodfoodcesi/auth-api/crypto"
	"github.com/goodfoodcesi/auth-api/domain/service"
	"github.com/goodfoodcesi/auth-api/infrastructure/database"
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
	//nolint:errcheck
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

	rabbit, err := rabbitmq.NewRabbitMQ(os.Getenv("RABBITMQ_URL"), logger)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	messagingService, err := service.NewMessagingService(rabbit, logger)
	if err != nil {
		log.Fatal(err)
	}

	tokenManager := jwt.NewTokenManager(
		os.Getenv("JWT_ACCESS_SECRET"),
		os.Getenv("JWT_REFRESH_SECRET"),
		time.Hour*24,    // Access token TTL
		time.Hour*24*30, // Refresh token TTL
	)

	userRepo := repository.NewUserRepository(db)
	passwordManager := crypto.NewPasswordManager(os.Getenv("PASSWORD_SECRET"))
	userService := service.NewUserService(userRepo, tokenManager, passwordManager, messagingService, logger)
	userHandler := handler.NewUserHandler(userService, logger)
	r := router.NewRouter(userHandler, logger, tokenManager)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Channel pour recevoir les signaux d'interruption
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Démarrer le serveur dans une goroutine
	go func() {
		logger.Info("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Attendre le signal d'arrêt
	<-stop
	logger.Info("Shutting down server...")

	// Créer un contexte avec timeout pour le shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Arrêter le serveur HTTP
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}
