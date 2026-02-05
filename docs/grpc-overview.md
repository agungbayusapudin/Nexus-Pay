# gRPC Implementation Overview

## File Structure
```
internal/api/grpc/
├── handler/
│   └── user_handler.go      # Business logic handlers
├── interceptor/
│   └── interceptor.go       # Middleware functions
└── server.go                # Server configuration & setup

cmd/server/
└── main.go                  # Application entry point
```

## Architecture Flow
```
Client Request → gRPC Server → Interceptors → Handler → Service → Repository → Database
                     ↓             ↓           ↓         ↓          ↓
                Configuration  Middleware   API Layer  Business   Data Layer
```

## Component Responsibilities

### 1. **main.go** - Application Bootstrap
- **Purpose**: Entry point dan dependency injection
- **Responsibilities**:
  - Load configuration
  - Initialize all connections (DB, RabbitMQ, gRPC clients)
  - Wire dependencies
  - Start server
  - Handle graceful shutdown

### 2. **server.go** - gRPC Server Configuration
- **Purpose**: Server setup dengan production-ready configuration
- **Responsibilities**:
  - TCP listener setup
  - Server options configuration (keepalive, limits, TLS)
  - Interceptor chain setup
  - Service registration
  - Lifecycle management (start/stop)

### 3. **interceptor.go** - Cross-Cutting Concerns
- **Purpose**: Middleware untuk semua gRPC requests
- **Responsibilities**:
  - Request/response logging
  - Panic recovery
  - Timeout handling
  - Validation (extensible)

### 4. **user_handler.go** - API Layer
- **Purpose**: gRPC service implementation
- **Responsibilities**:
  - Implement UserServiceServer interface
  - Request validation
  - Proto ↔ Domain model conversion
  - Error handling dan status code mapping
  - Call business logic layer

## Key Design Patterns

### 1. **Dependency Injection**
```go
// In main.go
userHandler := handler.NewUserHandler(userService)
server := grpcserver.NewServer(grpcConfig, userHandler)
```
- All dependencies injected via constructors
- Easy testing dengan mock dependencies
- Clear dependency graph

### 2. **Interceptor Chain**
```go
grpc.ChainUnaryInterceptor(
    interceptor.RecoveryInterceptor(),    // Outermost
    interceptor.LoggingInterceptor(),
    interceptor.ValidationInterceptor(),
    interceptor.TimeoutInterceptor(),     // Innermost
)
```
- Composable middleware
- Order matters (recovery first, timeout last)
- Separation of concerns

### 3. **Error Handling Strategy**
```go
// Domain errors → gRPC status codes
func (h *UserHandler) handleServiceError(err error) error {
    switch {
    case err == model.ErrUserNotFound:
        return status.Error(codes.NotFound, err.Error())
    case err == model.ErrEmailExists:
        return status.Error(codes.AlreadyExists, err.Error())
    default:
        return status.Error(codes.Internal, "Internal server error")
    }
}
```
- Proper gRPC status codes
- Hide internal errors dari clients
- Consistent error responses

### 4. **Graceful Shutdown**
```go
// Signal handling
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful stop
server.Stop() // Finish ongoing requests
```
- Handle OS signals
- Finish ongoing requests
- Cleanup resources

## Production Features

### Performance
- **Keepalive**: Connection management
- **Message Limits**: 4MB max message size
- **Concurrent Streams**: Configurable limits
- **Timeouts**: Request timeout handling

### Security
- **TLS Ready**: Certificate-based encryption
- **Input Validation**: Request validation
- **Error Hiding**: Internal errors hidden dari clients
- **Rate Limiting Ready**: Via interceptors

### Observability
- **Request Logging**: All requests logged dengan duration
- **Error Logging**: Detailed error information
- **Panic Recovery**: Graceful panic handling
- **Metrics Ready**: Structure untuk Prometheus integration

### Reliability
- **Graceful Shutdown**: Proper cleanup
- **Connection Pooling**: Database connection management
- **Retry Logic Ready**: Structure untuk retry mechanisms
- **Health Checks Ready**: Easy untuk add health endpoints

## Configuration Management
```go
// Environment-based configuration
cfg := config.Load()

// Configurable timeouts, ports, connections
grpcConfig := config.NewGRPCServerConfig(cfg)
```
- Environment variable based
- Default values provided
- Type-safe configuration structs

## Testing Strategy
```go
// Easy mocking karena dependency injection
mockService := &MockUserService{}
handler := handler.NewUserHandler(mockService)

// Test individual components
server := grpcserver.NewServer(config, handler)
```
- All dependencies injectable
- Each layer testable independently
- Mock-friendly interfaces

## Deployment Ready
- **Docker**: Container-friendly
- **Kubernetes**: Graceful shutdown compatible
- **Process Managers**: SIGTERM handling
- **Load Balancers**: Health check ready

## Next Steps untuk Complete Implementation
1. **Domain Layer**: Implement User model, repository interface, service interface
2. **Infrastructure Layer**: Implement repository dengan SQL queries
3. **Service Layer**: Implement business logic
4. **Database Migration**: Create users table
5. **Testing**: Unit tests untuk each layer
6. **Documentation**: API documentation dengan examples

Implementasi gRPC ini follows best practices untuk production microservices dengan proper separation of concerns, error handling, dan observability!
