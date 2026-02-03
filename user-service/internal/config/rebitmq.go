package config

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQConnection holds both connection and channel
type RabbitMQConnection struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

// NewRabbitMQConn creates new RabbitMQ connection and channel
func NewRabbitMQConn(cfg *Config) (*RabbitMQConnection, error) {
	// Build connection string from config
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQUser, cfg.RabbitMQPassword, cfg.RabbitMQHost, cfg.RabbitMQPort)

	// Create connection
	conn, err := amqp091.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close() // Close connection if channel creation fails
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQConnection{
		Conn:    conn,
		Channel: channel,
	}, nil
}

// Close closes both channel and connection
func (r *RabbitMQConnection) Close() error {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		return r.Conn.Close()
	}
	return nil
}
