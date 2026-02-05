package messagebroker

import (
	"time"
)

// Events yang dikirim oleh user-service
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
}

type UserUpdatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	UpdatedAt time.Time `json:"updated_at"`
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
}

type UserDeletedEvent struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
}

type UserStatusChangedEvent struct {
	UserID    string    `json:"user_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
}

// Events yang diterima dari service lain
type WalletCreatedEvent struct {
	UserID    string    `json:"user_id"`
	WalletID  string    `json:"wallet_id"`
	Balance   int64     `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
}

type TransactionCompletedEvent struct {
	UserID          string    `json:"user_id"`
	TransactionID   string    `json:"transaction_id"`
	Amount          int64     `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	CompletedAt     time.Time `json:"completed_at"`
	EventID         string    `json:"event_id"`
	EventType       string    `json:"event_type"`
}

type SecurityAlertEvent struct {
	UserID      string    `json:"user_id"`
	AlertType   string    `json:"alert_type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"created_at"`
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
}
