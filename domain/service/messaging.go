package service

import (
	"context"

	"github.com/goodfoodcesi/auth-api/domain/event"

	"github.com/goodfoodcesi/auth-api/infrastructure/messaging/rabbitmq"
	"go.uber.org/zap"
)

type MessagingService struct {
	rabbit *rabbitmq.RabbitMQ
	logger *zap.Logger
}

func NewMessagingService(rabbit *rabbitmq.RabbitMQ, logger *zap.Logger) (*MessagingService, error) {
	// Déclarer les exchanges
	exchanges := []rabbitmq.ExchangeConfig{
		{
			Name:    rabbitmq.UserExchange,
			Type:    rabbitmq.TopicExchange,
			Durable: true,
		},
		{
			Name:    rabbitmq.EmailExchange,
			Type:    rabbitmq.DirectExchange,
			Durable: true,
		},
	}

	// Déclarer les queues
	queues := []rabbitmq.QueueConfig{
		{
			Name:    rabbitmq.UserCreatedQueue,
			Durable: true,
		},
		{
			Name:    rabbitmq.WelcomeEmailQueue,
			Durable: true,
		},
	}

	// Déclarer les bindings
	bindings := []rabbitmq.BindingConfig{
		{
			Queue:      rabbitmq.UserCreatedQueue,
			Exchange:   rabbitmq.UserExchange,
			RoutingKey: rabbitmq.UserCreatedKey,
		},
		{
			Queue:      rabbitmq.WelcomeEmailQueue,
			Exchange:   rabbitmq.EmailExchange,
			RoutingKey: rabbitmq.WelcomeEmailKey,
		},
	}

	// Configurer RabbitMQ
	for _, ex := range exchanges {
		if err := rabbit.DeclareExchange(ex); err != nil {
			return nil, err
		}
	}

	for _, q := range queues {
		if err := rabbit.DeclareQueue(q); err != nil {
			return nil, err
		}
	}

	for _, b := range bindings {
		if err := rabbit.BindQueue(b); err != nil {
			return nil, err
		}
	}

	return &MessagingService{
		rabbit: rabbit,
		logger: logger,
	}, nil
}

func (s *MessagingService) PublishUserCreated(ctx context.Context, userCreatedEvent event.UserCreatedEvent) error {
	// Publier dans l'exchange utilisateur
	err := s.rabbit.Publish(ctx, rabbitmq.UserExchange, rabbitmq.UserCreatedKey, userCreatedEvent)
	if err != nil {
		s.logger.Error("failed to publish user created event", zap.Error(err))
		return err
	}

	// Publier dans l'exchange email pour l'email de bienvenue
	welcomeEmail := event.WelcomeEmailEvent{
		ID:        userCreatedEvent.ID,
		Email:     userCreatedEvent.Email,
		FirstName: userCreatedEvent.FirstName,
		LastName:  userCreatedEvent.LastName,
		Role:      userCreatedEvent.Role,
	}
	if err := s.rabbit.Publish(ctx, rabbitmq.EmailExchange, rabbitmq.WelcomeEmailKey, welcomeEmail); err != nil {
		s.logger.Error("failed to publish welcome email event", zap.Error(err))
	}

	return nil
}
