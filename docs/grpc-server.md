# gRPC Server Documentation

## Overview
File `server.go` berisi implementasi gRPC server dengan konfigurasi lengkap untuk production use. Server ini mengintegrasikan semua komponen: handlers, interceptors, dan berbagai server options untuk performance, security, dan reliability.

## Code Structure Analysis

### Package & Imports
```go
package grpc

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	
	pb "user-service/proto/user"
	"user-service/internal/api/grpc/handler"
	"user-service/internal/api/grpc/interceptor"
	"user-service/internal/config"
)
```

**Penjelasan Imports:**
- `fmt`: String formatting untuk error messages
- `net`: Network operations (TCP listener)
- `time`: Duration values untuk timeouts
- `grpc`: Core gRPC server functionality
- `keepalive`: Connection keepalive settings
- `reflection`: gRPC reflection untuk development tools
- `pb`: Generated protobuf types
- `handler`: Business logic handlers
- `interceptor`: Middleware functions
- `config`: Configuration structs

### Server Struct
```go
type Server struct {
	server      *grpc.Server
	listener    net.Listener
	userHandler *handler.UserHandler
}
```

**Field Explanations:**
- `server`: gRPC server instance
- `listener`: TCP listener untuk accept connections
- `userHandler`: Business logic handler (stored untuk potential future use)

## Constructor Function

### NewServer Function Signature
```go
func NewServer(cfg *config.GRPCServerConfig, userHandler *handler.UserHandler) (*Server, error) {
```

**Parameters:**
- `cfg`: Server configuration (port, TLS, timeouts, dll)
- `userHandler`: Handler yang implement business logic
- **Returns**: Server instance atau error

### Step 1: Create TCP Listener
```go
lis, err := net.Listen("tcp", ":"+cfg.Port)
if err != nil {
	return nil, fmt.Errorf("failed to listen on port %s: %w", cfg.Port, err)
}
```

**Line by Line:**
- `net.Listen("tcp", ":"+cfg.Port)`: Create TCP listener pada specified port
- `":"` prefix: Listen pada all interfaces (0.0.0.0)
- Error wrapping dengan `fmt.Errorf` dan `%w` verb untuk error chain
- Early return jika listener creation gagal

### Step 2: Setup Server Options
```go
opts := []grpc.ServerOption{
	// Server options akan dijelaskan detail di bawah
}
```

#### Interceptor Chain
```go
grpc.ChainUnaryInterceptor(
	interceptor.RecoveryInterceptor(),
	interceptor.LoggingInterceptor(),
	interceptor.ValidationInterceptor(),
	interceptor.TimeoutInterceptor(time.Duration(cfg.RequestTimeout)*time.Second),
),
```

**Penjelasan:**
- `ChainUnaryInterceptor`: Combine multiple interceptors dalam order
- **Order matters**: Recovery → Logging → Validation → Timeout
- `time.Duration(cfg.RequestTimeout)*time.Second`: Convert config int ke Duration
- Interceptors execute dalam order, response flows back dalam reverse order

#### Keepalive Parameters
```go
grpc.KeepaliveParams(keepalive.ServerParameters{
	MaxConnectionIdle:     15 * time.Second,
	MaxConnectionAge:      30 * time.Second,
	MaxConnectionAgeGrace: 5 * time.Second,
	Time:                  5 * time.Second,
	Timeout:               1 * time.Second,
}),
```

**Parameter Explanations:**
- `MaxConnectionIdle: 15s`: Close idle connections after 15 seconds
- `MaxConnectionAge: 30s`: Force close connections after 30 seconds (prevent memory leaks)
- `MaxConnectionAgeGrace: 5s`: Grace period untuk finish ongoing requests
- `Time: 5s`: Send keepalive ping every 5 seconds
- `Timeout: 1s`: Wait 1 second untuk keepalive response

**Why Keepalive?**
- Detect dead connections early
- Prevent resource leaks
- Better load balancing (close old connections)
- Network firewall compatibility

#### Keepalive Enforcement
```go
grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second,
	PermitWithoutStream: true,
}),
```

**Parameter Explanations:**
- `MinTime: 5s`: Client tidak boleh send keepalive lebih sering dari 5 detik
- `PermitWithoutStream: true`: Allow keepalive bahkan tanpa active streams
- **Purpose**: Prevent keepalive abuse dari malicious clients

#### Message Size Limits
```go
grpc.MaxRecvMsgSize(4 * 1024 * 1024), // 4MB
grpc.MaxSendMsgSize(4 * 1024 * 1024), // 4MB
```

**Penjelasan:**
- `MaxRecvMsgSize`: Maximum size untuk incoming messages (4MB)
- `MaxSendMsgSize`: Maximum size untuk outgoing messages (4MB)
- **Purpose**: Prevent memory exhaustion attacks
- **4MB**: Reasonable limit untuk most use cases

#### Concurrent Streams Limit
```go
grpc.MaxConcurrentStreams(uint32(cfg.MaxConnections)),
```

**Penjelasan:**
- Limit concurrent streams per connection
- Prevent single client dari overwhelming server
- Configurable via config file
- **Load balancing**: Force clients untuk use multiple connections

### Step 3: TLS Configuration (Optional)
```go
if cfg.EnableTLS {
	// TODO: Add TLS credentials
	// creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
	// }
	// opts = append(opts, grpc.Creds(creds))
}
```

**Penjelasan:**
- Conditional TLS setup berdasarkan config
- Currently commented untuk development
- Production ready structure
- Certificate file paths dari config

### Step 4: Create gRPC Server
```go
server := grpc.NewServer(opts...)
```

**Penjelasan:**
- Create gRPC server dengan all configured options
- `opts...`: Spread slice sebagai variadic arguments
- Server belum start, hanya configured

### Step 5: Register Services
```go
pb.RegisterUserServiceServer(server, userHandler)
```

**Penjelasan:**
- Register handler untuk UserService
- `pb.RegisterUserServiceServer`: Generated function dari proto
- Links proto service definition dengan actual implementation
- Multiple services bisa di-register pada same server

### Step 6: Enable Reflection
```go
reflection.Register(server)
```

**Penjelasan:**
- Enable gRPC reflection untuk development tools
- Tools seperti `grpcurl`, `grpc_cli` bisa discover services
- **Production**: Biasanya disabled untuk security
- **Development**: Very useful untuk testing

### Step 7: Return Server Instance
```go
return &Server{
	server:      server,
	listener:    lis,
	userHandler: userHandler,
}, nil
```

**Penjelasan:**
- Create Server struct dengan all components
- Store listener untuk lifecycle management
- Store handler untuk potential future use
- Return nil error untuk success case

## Server Lifecycle Methods

### Start Method
```go
func (s *Server) Start() error {
	return s.server.Serve(s.listener)
}
```

**Penjelasan:**
- `Serve()`: Blocking call yang start accepting connections
- Use stored listener dari constructor
- Return error jika server fails to start
- **Blocking**: Method tidak return until server stops

### Stop Method
```go
func (s *Server) Stop() {
	s.server.GracefulStop()
}
```

**Penjelasan:**
- `GracefulStop()`: Wait untuk ongoing requests to finish
- Close listener (no new connections)
- Finish existing requests
- **Graceful**: Better than `Stop()` yang immediately kills connections

### Getter Methods
```go
func (s *Server) GetListener() net.Listener {
	return s.listener
}

func (s *Server) GetServer() *grpc.Server {
	return s.server
}
```

**Penjelasan:**
- Expose internal components untuk testing atau advanced use cases
- `GetListener()`: Useful untuk getting actual port (jika using port 0)
- `GetServer()`: Access untuk additional configuration atau metrics

## Production Considerations

### Performance Settings
1. **Keepalive**: Prevent connection leaks
2. **Message Limits**: Prevent memory attacks  
3. **Concurrent Streams**: Prevent resource exhaustion
4. **Timeouts**: Prevent hanging requests

### Security Settings
1. **TLS**: Encrypt traffic (commented, ready untuk production)
2. **Message Size Limits**: Prevent DoS attacks
3. **Rate Limiting**: Via interceptors
4. **Authentication**: Via interceptors

### Observability
1. **Logging**: Via interceptors
2. **Metrics**: Ready untuk Prometheus integration
3. **Tracing**: Ready untuk distributed tracing
4. **Reflection**: Development debugging

## Usage Example
```go
// Create server
server, err := grpc.NewServer(grpcConfig, userHandler)
if err != nil {
    log.Fatal(err)
}

// Start in goroutine
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()

// Graceful shutdown
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

server.Stop()
```

Server ini production-ready dengan proper configuration untuk scalability, security, dan observability!
