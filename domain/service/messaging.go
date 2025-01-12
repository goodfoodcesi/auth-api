package service

import (
	"context"

	"github.com/goodfoodcesi/api-utils-go/pkg/event"

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
			Name:       rabbitmq.ClientExchange,
			Type:       rabbitmq.FanoutExchange,
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
		},
	}

	for _, ex := range exchanges {
		if err := rabbit.DeclareExchange(ex); err != nil {
			return nil, err
		}
	}

	// Déclarer les queues
	queues := []rabbitmq.QueueConfig{
		{
			Name:       rabbitmq.ClientCreatedQueueNotificationAPI,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
		},
		{
			Name:       rabbitmq.ClientCreatedQueueClientAPI,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
		},
	}

	for _, q := range queues {
		if err := rabbit.DeclareQueue(q); err != nil {
			return nil, err
		}
	}

	// Déclarer les bindings
	bindings := []rabbitmq.BindingConfig{
		{
			Queue:    rabbitmq.ClientCreatedQueueNotificationAPI,
			Exchange: rabbitmq.ClientExchange,
			NoWait:   false,
		},
		{
			Queue:    rabbitmq.ClientCreatedQueueClientAPI,
			Exchange: rabbitmq.ClientExchange,
			NoWait:   false,
		},
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
	err := s.rabbit.Publish(ctx, rabbitmq.ClientExchange, "", userCreatedEvent)
	if err != nil {
		s.logger.Error("failed to publish user created event", zap.Error(err))
		return err
	}

	s.logger.Info("user created event published")

	return nil
}
