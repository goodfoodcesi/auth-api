package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"sync"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchanges map[string]ExchangeConfig
	queues    map[string]QueueConfig
	bindings  []BindingConfig
	logger    *zap.Logger
	mu        sync.RWMutex
}

func NewRabbitMQ(url string, logger *zap.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	logger.Info("Connected to RabbitMQ")

	return &RabbitMQ{
		conn:      conn,
		channel:   ch,
		exchanges: make(map[string]ExchangeConfig),
		queues:    make(map[string]QueueConfig),
		logger:    logger,
	}, nil
}

func (r *RabbitMQ) DeclareExchange(config ExchangeConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.channel.ExchangeDeclare(
		config.Name,
		string(config.Type),
		config.Durable,
		config.AutoDelete,
		config.Internal,
		config.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	r.exchanges[config.Name] = config
	return nil
}

func (r *RabbitMQ) DeclareQueue(config QueueConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.channel.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	r.queues[config.Name] = config
	return nil
}

func (r *RabbitMQ) BindQueue(config BindingConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.channel.QueueBind(
		config.Queue,
		config.RoutingKey,
		config.Exchange,
		config.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	r.bindings = append(r.bindings, config)
	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, exchange, routingKey string, message interface{}) error {
	if err := r.ensureConnection(); err != nil {
		return fmt.Errorf("failed to ensure RabbitMQ connection: %w", err)
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) ensureConnection() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil || r.conn.IsClosed() {
		var err error
		r.conn, err = amqp.Dial(r.conn.LocalAddr().String())
		if err != nil {
			return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
		}

		r.channel, err = r.conn.Channel()
		if err != nil {
			r.conn.Close()
			return fmt.Errorf("failed to reopen channel: %w", err)
		}
	}
	return nil
}

func (r *RabbitMQ) Consume(queueName string, handler func([]byte) error) error {
	msgs, err := r.channel.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				r.logger.Error("failed to process message",
					zap.Error(err),
					zap.String("queue", queueName),
				)
				err := msg.Nack(false, true)
				if err != nil {
					r.logger.Error("failed to nack message", zap.Error(err))
					return
				} // requeue message
			} else {
				err := msg.Ack(false)
				if err != nil {
					r.logger.Error("failed to ack message", zap.Error(err))
					return
				}
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.channel.Close(); err != nil {
		r.logger.Error("failed to close channel", zap.Error(err))
	}
	return r.conn.Close()
}
