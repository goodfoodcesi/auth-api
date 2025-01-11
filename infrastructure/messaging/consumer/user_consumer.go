package consumer

import (
	"encoding/json"
	"github.com/goodfoodcesi/auth-api/domain/event"
	"go.uber.org/zap"
)

type UserConsumer struct {
	logger *zap.Logger
}

func NewUserConsumer(logger *zap.Logger) *UserConsumer {
	return &UserConsumer{
		logger: logger,
	}
}

func (c *UserConsumer) HandleUserCreated(data []byte) error {
	var userCreatedEvent event.UserCreatedEvent
	if err := json.Unmarshal(data, &userCreatedEvent); err != nil {
		return err
	}

	c.logger.Info("handling user created event",
		zap.String("user_id", userCreatedEvent.ID.String()),
		zap.String("email", userCreatedEvent.Email),
	)

	// Logique de traitement...
	return nil
}
