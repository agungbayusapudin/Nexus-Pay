package config

import (
	"os"
	"strconv"
)

// Config menampung semua variabel environment
type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	RabbitMQHost         string
	RabbitMQPort         string
	RabbitMQUser         string
	RabbitMQPassword     string
	GRPCPort             string
	WalletServiceURL     string
	TransactionServiceURL string
}

// Load akan membaca data dari .env atau environment variable
func Load() *Config {
	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "password"),
		DBName:               getEnv("DB_NAME", "db_transaction"),
		RabbitMQHost:         getEnv("RABBITMQ_HOST", "localhost"),
		RabbitMQPort:         getEnv("RABBITMQ_PORT", "5672"),
		RabbitMQUser:         getEnv("RABBITMQ_USER", "guest"),
		RabbitMQPassword:     getEnv("RABBITMQ_PASSWORD", "guest"),
		GRPCPort:             getEnv("GRPC_PORT", "50051"),
		WalletServiceURL:     getEnv("WALLET_SERVICE_URL", "localhost:50052"),
		TransactionServiceURL: getEnv("TRANSACTION_SERVICE_URL", "localhost:50053"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		return value == "true"
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
