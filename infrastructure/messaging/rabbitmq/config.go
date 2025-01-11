package rabbitmq

type ExchangeType string

const (
	DirectExchange ExchangeType = "direct"
	FanoutExchange ExchangeType = "fanout"
	TopicExchange  ExchangeType = "topic"
)

type ExchangeConfig struct {
	Name       string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
}

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
}

type BindingConfig struct {
	Queue      string
	Exchange   string
	RoutingKey string
	NoWait     bool
}
