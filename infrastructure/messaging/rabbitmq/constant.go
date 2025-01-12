package rabbitmq

const (
	// Exchanges
	ClientExchange = "client.events"

	// Queues
	ClientCreatedQueueNotificationAPI = "client.created.notification-api"
	ClientCreatedQueueClientAPI       = "client.created.client-api"

	// Routing Keys (not needed for fanout, but keeping for reference)
	ClientCreatedKey = "client.created"
)
