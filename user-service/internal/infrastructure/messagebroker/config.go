package messagebroker

// Exchanges
const (
	UserExchange        = "user.events"
	WalletExchange      = "wallet.events"
	TransactionExchange = "transaction.events"
	SecurityExchange    = "security.events"
)

// Queues for user-service
const (
	UserCreatedQueue       = "user.created"
	UserUpdatedQueue       = "user.updated"
	UserDeletedQueue       = "user.deleted"
	UserStatusChangedQueue = "user.status.changed"
	WalletEventsQueue      = "user.wallet.events"
	TransactionEventsQueue = "user.transaction.events"
	SecurityEventsQueue    = "user.security.events"
)

// Routing Keys
const (
	// User events routing keys
	UserCreatedRoutingKey       = "user.created"
	UserUpdatedRoutingKey       = "user.updated"
	UserDeletedRoutingKey       = "user.deleted"
	UserStatusChangedRoutingKey = "user.status.changed"
	
	// External events routing keys
	WalletCreatedRoutingKey       = "wallet.created"
	TransactionCompletedRoutingKey = "transaction.completed"
	SecurityAlertRoutingKey       = "security.alert"
)

// Event Types
const (
	EventTypeUserCreated       = "user.created"
	EventTypeUserUpdated       = "user.updated"
	EventTypeUserDeleted       = "user.deleted"
	EventTypeUserStatusChanged = "user.status.changed"
	EventTypeWalletCreated     = "wallet.created"
	EventTypeTransactionCompleted = "transaction.completed"
	EventTypeSecurityAlert     = "security.alert"
)
