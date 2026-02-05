package messagebroker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"user-service/internal/config"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

// Publisher interface for publishing events
type Publisher interface {
	PublishUserCreated(ctx context.Context, event *UserCreatedEvent) error
	PublishUserUpdated(ctx context.Context, event *UserUpdatedEvent) error
	PublishUserDeleted(ctx context.Context, event *UserDeletedEvent) error
	PublishUserStatusChanged(ctx context.Context, event *UserStatusChangedEvent) error
	Close() error
}

// publisher implements Publisher interface
type publisher struct {
	rabbitmq *config.RabbitMQConnection
}

// NewPublisher creates new publisher
func NewPublisher(rabbitmq *config.RabbitMQConnection) (Publisher, error) {
	p := &publisher{rabbitmq: rabbitmq}

	// Setup exchanges and queues
	if err := p.setupInfrastructure(); err != nil {
		return nil, fmt.Errorf("failed to setup publisher infrastructure: %w", err)
	}

	return p, nil
}

// setupInfrastructure declares exchanges and queues
func (p *publisher) setupInfrastructure() error {
	// Declare user exchange
	err := p.rabbitmq.Channel.ExchangeDeclare(
		UserExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare user exchange: %w", err)
	}

	// Declare queues
	queues := []string{
		UserCreatedQueue,
		UserUpdatedQueue,
		UserDeletedQueue,
		UserStatusChangedQueue,
	}

	for _, queueName := range queues {
		_, err := p.rabbitmq.Channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}
	}

	// Bind queues to exchange
	bindings := map[string]string{
		UserCreatedQueue:       UserCreatedRoutingKey,
		UserUpdatedQueue:       UserUpdatedRoutingKey,
		UserDeletedQueue:       UserDeletedRoutingKey,
		UserStatusChangedQueue: UserStatusChangedRoutingKey,
	}

	for queueName, routingKey := range bindings {
		err := p.rabbitmq.Channel.QueueBind(
			queueName,    // queue name
			routingKey,   // routing key
			UserExchange, // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", queueName, err)
		}
	}

	return nil
}

// PublishUserCreated publishes user created event
func (p *publisher) PublishUserCreated(ctx context.Context, event *UserCreatedEvent) error {
	event.EventID = uuid.New().String()
	event.EventType = EventTypeUserCreated

	return p.publishEvent(ctx, UserExchange, UserCreatedRoutingKey, event)
}

// PublishUserUpdated publishes user updated event
func (p *publisher) PublishUserUpdated(ctx context.Context, event *UserUpdatedEvent) error {
	event.EventID = uuid.New().String()
	event.EventType = EventTypeUserUpdated

	return p.publishEvent(ctx, UserExchange, UserUpdatedRoutingKey, event)
}

// PublishUserDeleted publishes user deleted event
func (p *publisher) PublishUserDeleted(ctx context.Context, event *UserDeletedEvent) error {
	event.EventID = uuid.New().String()
	event.EventType = EventTypeUserDeleted

	return p.publishEvent(ctx, UserExchange, UserDeletedRoutingKey, event)
}

// PublishUserStatusChanged publishes user status changed event
func (p *publisher) PublishUserStatusChanged(ctx context.Context, event *UserStatusChangedEvent) error {
	event.EventID = uuid.New().String()
	event.EventType = EventTypeUserStatusChanged

	return p.publishEvent(ctx, UserExchange, UserStatusChangedRoutingKey, event)
}

// publishEvent publishes event to RabbitMQ
func (p *publisher) publishEvent(ctx context.Context, exchange, routingKey string, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.rabbitmq.Channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent, // make message persistent
			Timestamp:    time.Now(),
			Body:         body,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Close closes the publisher
func (p *publisher) Close() error {
	// Publisher doesn't own the connection, so don't close it
	return nil
}
