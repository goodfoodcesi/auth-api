// internal/infrastructure/messaging/rabbitmq/constants.go
package rabbitmq

const (
	// Exchanges
	UserExchange  = "user.events"
	EmailExchange = "email.events"

	// Queues
	UserCreatedQueue  = "user.created"
	UserUpdatedQueue  = "user.updated"
	UserDeletedQueue  = "user.deleted"
	WelcomeEmailQueue = "email.welcome"

	// Routing Keys
	UserCreatedKey  = "user.created"
	UserUpdatedKey  = "user.updated"
	UserDeletedKey  = "user.deleted"
	WelcomeEmailKey = "email.welcome"
)
