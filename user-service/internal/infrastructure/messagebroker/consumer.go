package messagebroker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"user-service/internal/config"
)

// Consumer interface for consuming events
type Consumer interface {
	ConsumeWalletEvents(ctx context.Context, handler WalletEventHandler) error
	ConsumeTransactionEvents(ctx context.Context, handler TransactionEventHandler) error
	ConsumeSecurityEvents(ctx context.Context, handler SecurityEventHandler) error
	Close() error
}

// Event handlers
type WalletEventHandler interface {
	HandleWalletCreated(ctx context.Context, event *WalletCreatedEvent) error
}

type TransactionEventHandler interface {
	HandleTransactionCompleted(ctx context.Context, event *TransactionCompletedEvent) error
}

type SecurityEventHandler interface {
	HandleSecurityAlert(ctx context.Context, event *SecurityAlertEvent) error
}

// consumer implements Consumer interface
type consumer struct {
	rabbitmq *config.RabbitMQConnection
}

// NewConsumer creates new consumer
func NewConsumer(rabbitmq *config.RabbitMQConnection) (Consumer, error) {
	c := &consumer{rabbitmq: rabbitmq}
	
	// Setup consumer infrastructure
	if err := c.setupInfrastructure(); err != nil {
		return nil, fmt.Errorf("failed to setup consumer infrastructure: %w", err)
	}
	
	return c, nil
}

// setupInfrastructure declares exchanges and queues for consuming
func (c *consumer) setupInfrastructure() error {
	// Declare external exchanges
	exchanges := []string{WalletExchange, TransactionExchange, SecurityExchange}
	
	for _, exchange := range exchanges {
		err := c.rabbitmq.Channel.ExchangeDeclare(
			exchange, // name
			"topic",  // type
			true,     // durable
			false,    // auto-deleted
			false,    // internal
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange, err)
		}
	}
	
	// Declare consumer queues
	queues := []string{
		WalletEventsQueue,
		TransactionEventsQueue,
		SecurityEventsQueue,
	}
	
	for _, queueName := range queues {
		_, err := c.rabbitmq.Channel.QueueDeclare(
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
	
	// Bind queues to exchanges
	bindings := map[string]map[string]string{
		WalletExchange: {
			WalletEventsQueue: WalletCreatedRoutingKey,
		},
		TransactionExchange: {
			TransactionEventsQueue: TransactionCompletedRoutingKey,
		},
		SecurityExchange: {
			SecurityEventsQueue: SecurityAlertRoutingKey,
		},
	}
	
	for exchange, queueBindings := range bindings {
		for queueName, routingKey := range queueBindings {
			err := c.rabbitmq.Channel.QueueBind(
				queueName,  // queue name
				routingKey, // routing key
				exchange,   // exchange
				false,      // no-wait
				nil,        // arguments
			)
			if err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchange, err)
			}
		}
	}
	
	return nil
}

// ConsumeWalletEvents consumes wallet events
func (c *consumer) ConsumeWalletEvents(ctx context.Context, handler WalletEventHandler) error {
	msgs, err := c.rabbitmq.Channel.Consume(
		WalletEventsQueue, // queue
		"",                // consumer
		false,             // auto-ack (manual ack for reliability)
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register wallet consumer: %w", err)
	}
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				
				if err := c.handleWalletMessage(ctx, msg, handler); err != nil {
					log.Printf("Error handling wallet message: %v", err)
					msg.Nack(false, true) // Requeue message
				} else {
					msg.Ack(false) // Acknowledge successful processing
				}
			}
		}
	}()
	
	return nil
}

// ConsumeTransactionEvents consumes transaction events
func (c *consumer) ConsumeTransactionEvents(ctx context.Context, handler TransactionEventHandler) error {
	msgs, err := c.rabbitmq.Channel.Consume(
		TransactionEventsQueue, // queue
		"",                     // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register transaction consumer: %w", err)
	}
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				
				if err := c.handleTransactionMessage(ctx, msg, handler); err != nil {
					log.Printf("Error handling transaction message: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()
	
	return nil
}

// ConsumeSecurityEvents consumes security events
func (c *consumer) ConsumeSecurityEvents(ctx context.Context, handler SecurityEventHandler) error {
	msgs, err := c.rabbitmq.Channel.Consume(
		SecurityEventsQueue, // queue
		"",                  // consumer
		false,               // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		return fmt.Errorf("failed to register security consumer: %w", err)
	}
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				
				if err := c.handleSecurityMessage(ctx, msg, handler); err != nil {
					log.Printf("Error handling security message: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()
	
	return nil
}

// handleWalletMessage handles wallet event messages
func (c *consumer) handleWalletMessage(ctx context.Context, msg amqp091.Delivery, handler WalletEventHandler) error {
	switch msg.RoutingKey {
	case WalletCreatedRoutingKey:
		var event WalletCreatedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			return fmt.Errorf("failed to unmarshal wallet created event: %w", err)
		}
		return handler.HandleWalletCreated(ctx, &event)
	default:
		log.Printf("Unknown wallet routing key: %s", msg.RoutingKey)
		return nil // Don't requeue unknown messages
	}
}

// handleTransactionMessage handles transaction event messages
func (c *consumer) handleTransactionMessage(ctx context.Context, msg amqp091.Delivery, handler TransactionEventHandler) error {
	switch msg.RoutingKey {
	case TransactionCompletedRoutingKey:
		var event TransactionCompletedEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			return fmt.Errorf("failed to unmarshal transaction completed event: %w", err)
		}
		return handler.HandleTransactionCompleted(ctx, &event)
	default:
		log.Printf("Unknown transaction routing key: %s", msg.RoutingKey)
		return nil
	}
}

// handleSecurityMessage handles security event messages
func (c *consumer) handleSecurityMessage(ctx context.Context, msg amqp091.Delivery, handler SecurityEventHandler) error {
	switch msg.RoutingKey {
	case SecurityAlertRoutingKey:
		var event SecurityAlertEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			return fmt.Errorf("failed to unmarshal security alert event: %w", err)
		}
		return handler.HandleSecurityAlert(ctx, &event)
	default:
		log.Printf("Unknown security routing key: %s", msg.RoutingKey)
		return nil
	}
}

// Close closes the consumer
func (c *consumer) Close() error {
	// Consumer doesn't own the connection, so don't close it
	return nil
}
