package config

// gRPC Server configuration for user-service
type GRPCServerConfig struct {
	Port            string
	EnableTLS       bool
	CertFile        string
	KeyFile         string
	MaxConnections  int
	RequestTimeout  int // seconds
}

// NewGRPCServerConfig creates gRPC server configuration
func NewGRPCServerConfig(cfg *Config) *GRPCServerConfig {
	return &GRPCServerConfig{
		Port:            cfg.GRPCPort,
		EnableTLS:       getEnvBool("GRPC_ENABLE_TLS", false),
		CertFile:        getEnv("GRPC_CERT_FILE", ""),
		KeyFile:         getEnv("GRPC_KEY_FILE", ""),
		MaxConnections:  getEnvInt("GRPC_MAX_CONNECTIONS", 100),
		RequestTimeout:  getEnvInt("GRPC_REQUEST_TIMEOUT", 30),
	}
}
