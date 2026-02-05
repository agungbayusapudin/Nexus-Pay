# Main Server Documentation

## Overview
File `main.go` adalah entry point aplikasi yang bertanggung jawab untuk:
- Bootstrap semua dependencies
- Initialize connections (database, RabbitMQ, gRPC clients)
- Setup dan start gRPC server
- Handle graceful shutdown

## Code Structure Analysis

### Package & Imports
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"user-service/internal/config"
	grpcserver "user-service/internal/api/grpc"
	"user-service/internal/api/grpc/handler"
	"user-service/internal/infrastructure/grpclient"
	// Import lain akan ditambah setelah layer dibuat
)
```

**Penjelasan Imports:**
- `context`: Untuk context handling dan cancellation
- `log`: Logging untuk startup process
- `os`: Operating system interface
- `os/signal`: Signal handling untuk graceful shutdown
- `syscall`: System call constants (SIGINT, SIGTERM)
- `config`: Configuration loading
- `grpcserver`: Alias untuk avoid naming conflict dengan grpc package
- `handler`: gRPC handlers
- `grpclient`: Client connections ke service lain

### Main Function Structure
```go
func main() {
	// 1. Configuration Loading
	// 2. Database Connection
	// 3. RabbitMQ Connection  
	// 4. gRPC Client Connections
	// 5. Repository Layer (commented)
	// 6. Service Layer (commented)
	// 7. Handler Layer
	// 8. gRPC Server Setup
	// 9. Server Start
	// 10. Graceful Shutdown
}
```

## Step-by-Step Implementation

### Step 1: Configuration Loading
```go
cfg := config.Load()
log.Printf("Starting user-service on port %s", cfg.GRPCPort)
```

**Line by Line:**
- `config.Load()`: Load configuration dari environment variables atau defaults
- `log.Printf()`: Log startup message dengan port information
- **Early logging**: Important untuk debugging startup issues

### Step 2: Database Connection
```go
db, err := config.NewPostgresConn(cfg)
if err != nil {
	log.Fatalf("Failed to connect to database: %v", err)
}
defer db.Close()
log.Println("Database connected successfully")
```

**Line by Line:**
- `config.NewPostgresConn(cfg)`: Create PostgreSQL connection dengan config
- `log.Fatalf()`: Fatal error jika database connection gagal - app cannot continue
- `defer db.Close()`: Ensure database connection closed saat app exit
- Success logging untuk confirmation

**Why Fatal on DB Error?**
- Database adalah critical dependency
- App tidak bisa function tanpa database
- Fail fast principle - better than partial functionality

### Step 3: RabbitMQ Connection
```go
rabbitmq, err := config.NewRabbitMQConn(cfg)
if err != nil {
	log.Fatalf("Failed to connect to RabbitMQ: %v", err)
}
defer rabbitmq.Close()
log.Println("RabbitMQ connected successfully")
```

**Line by Line:**
- `config.NewRabbitMQConn(cfg)`: Create RabbitMQ connection dan channel
- Fatal error handling - RabbitMQ critical untuk event-driven architecture
- `defer rabbitmq.Close()`: Cleanup connection dan channel
- Success confirmation logging

### Step 4: gRPC Client Connections
```go
grpcClients, err := grpclient.NewGRPCClients(cfg)
if err != nil {
	log.Fatalf("Failed to initialize gRPC clients: %v", err)
}
defer grpcClients.Close()
log.Println("gRPC clients initialized successfully")
```

**Line by Line:**
- `grpclient.NewGRPCClients(cfg)`: Initialize connections ke other services
- Fatal error - app needs to communicate dengan other services
- `defer grpcClients.Close()`: Cleanup all client connections
- Success logging untuk monitoring

### Step 5-6: Repository & Service Layers (Commented)
```go
// 5. Initialize Repository Layer
// userRepo := repository.NewUserRepository(db)

// 6. Initialize Service Layer
// userService := services.NewUserService(userRepo, rabbitmq, grpcClients)
```

**Penjelasan:**
- Commented karena layers belum diimplementasikan
- Shows dependency injection pattern
- Repository depends on database
- Service depends on repository, rabbitmq, dan grpcClients

### Step 7: Handler Layer
```go
// 7. Initialize gRPC Handler
// userHandler := handler.NewUserHandler(userService)

// Temporary: Create empty handler for now
userHandler := handler.NewUserHandler(nil)
```

**Line by Line:**
- Commented line shows proper dependency injection
- Temporary implementation dengan nil service untuk testing
- **Production**: Uncomment first line, remove temporary line

### Step 8: gRPC Server Setup
```go
grpcConfig := config.NewGRPCServerConfig(cfg)
server, err := grpcserver.NewServer(grpcConfig, userHandler)
if err != nil {
	log.Fatalf("Failed to create gRPC server: %v", err)
}
```

**Line by Line:**
- `config.NewGRPCServerConfig(cfg)`: Create gRPC-specific configuration
- `grpcserver.NewServer()`: Create configured gRPC server dengan handler
- Fatal error - server creation failure adalah critical

### Step 9: Server Start (Goroutine)
```go
go func() {
	log.Printf("gRPC server listening on port %s", cfg.GRPCPort)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}()
```

**Line by Line:**
- `go func()`: Start server dalam separate goroutine
- `server.Start()`: Blocking call untuk start accepting connections
- **Goroutine**: Allows main thread untuk handle shutdown signals
- Fatal error jika server fails to start

**Why Goroutine?**
- `server.Start()` adalah blocking call
- Main thread perlu available untuk signal handling
- Concurrent execution pattern

### Step 10: Graceful Shutdown
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down server...")

server.Stop()
log.Println("Server stopped gracefully")
```

**Line by Line Analysis:**

#### Signal Channel Setup
```go
quit := make(chan os.Signal, 1)
```
- Create buffered channel dengan capacity 1
- Buffer prevents signal loss jika multiple signals sent
- Channel untuk receive OS signals

#### Signal Registration
```go
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
```
- Register channel untuk receive specific signals
- `SIGINT`: Ctrl+C (interrupt signal)
- `SIGTERM`: Termination signal dari process manager
- **Production**: Process managers send SIGTERM untuk graceful shutdown

#### Wait for Signal
```go
<-quit
```
- Blocking receive dari signal channel
- Main thread waits here until signal received
- **Blocking**: App continues running until signal

#### Shutdown Process
```go
log.Println("Shutting down server...")
server.Stop()
log.Println("Server stopped gracefully")
```
- Log shutdown initiation
- `server.Stop()`: Graceful shutdown - finish ongoing requests
- Log completion confirmation

## Dependency Injection Pattern

### Current Structure (Temporary)
```
main() → config → db, rabbitmq, grpcClients → handler(nil) → server
```

### Production Structure (When Complete)
```
main() → config → db, rabbitmq, grpcClients → repository → service → handler → server
```

**Benefits:**
- **Testability**: Easy untuk mock dependencies
- **Flexibility**: Easy untuk swap implementations
- **Separation of Concerns**: Each layer has single responsibility
- **Configuration**: All dependencies configured dalam main()

## Error Handling Strategy

### Fatal Errors (App Cannot Continue)
- Database connection failure
- RabbitMQ connection failure
- gRPC clients initialization failure
- Server creation failure
- Server start failure

### Non-Fatal Errors (Log and Continue)
- Individual request failures (handled dalam handlers)
- Temporary connection issues (handled dengan retry logic)

## Production Considerations

### Logging
- Startup progress logging
- Success confirmations
- Error details untuk debugging
- **Production**: Use structured logging (JSON format)

### Monitoring
- Health checks untuk all dependencies
- Metrics collection
- Distributed tracing
- **Production**: Add health check endpoint

### Configuration
- Environment-based configuration
- Secrets management
- **Production**: Use proper secret management (Vault, K8s secrets)

### Deployment
- Container-ready (Docker)
- Kubernetes-ready (graceful shutdown)
- Process manager compatible (systemd)

## Usage Example

### Development
```bash
# Set environment variables
export DB_HOST=localhost
export RABBITMQ_HOST=localhost
export GRPC_PORT=50051

# Run application
go run cmd/server/main.go
```

### Production
```bash
# Docker
docker run -e DB_HOST=prod-db -e RABBITMQ_HOST=prod-mq user-service

# Kubernetes
kubectl apply -f deployment.yaml
```

### Graceful Shutdown
```bash
# Send SIGTERM
kill -TERM <pid>

# Or Ctrl+C untuk SIGINT
```

Main function ini implements best practices untuk production-ready microservice dengan proper dependency management, error handling, dan graceful shutdown!
