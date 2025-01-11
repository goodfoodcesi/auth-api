package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goodfoodcesi/auth-api/infrastructure/database"
	"github.com/goodfoodcesi/auth-api/infrastructure/logger"
	"github.com/goodfoodcesi/auth-api/infrastructure/messaging/consumer"
	"github.com/goodfoodcesi/auth-api/infrastructure/messaging/rabbitmq"
	"go.uber.org/zap"
)

func main() {

	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			log.Fatal("Failed to sync logger:", err)
		}
	}(logger)

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	userConsumer := consumer.NewUserConsumer(logger)

	go func() {
		err = rabbit.Consume(rabbitmq.UserCreatedQueue, userConsumer.HandleUserCreated)
		if err != nil {
			logger.Error("failed to consume user created event", zap.Error(err))
			stop <- syscall.SIGTERM
			return
		}
	}()

	logger.Info("Consumer started successfully")

	<-stop
	logger.Info("Shutting down consumer...")
}
