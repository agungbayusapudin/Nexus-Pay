# gRPC Interceptors Documentation

## Overview
File `interceptor.go` berisi middleware functions yang dijalankan sebelum dan sesudah setiap gRPC request. Interceptors adalah cara untuk menambahkan cross-cutting concerns seperti logging, authentication, validation, dll tanpa mengubah business logic.

## Interceptor Pattern
```go
type grpc.UnaryServerInterceptor func(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error)
```

**Parameters:**
- `ctx`: Request context
- `req`: Request message (interface{} karena bisa any proto message)
- `info`: Metadata tentang method yang dipanggil
- `handler`: Next handler dalam chain (bisa interceptor lain atau actual handler)

## Code Analysis

### Package & Imports
```go
package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
```

**Penjelasan Imports:**
- `context`: Untuk context handling
- `log`: Untuk logging operations
- `time`: Untuk duration measurement
- `grpc`: Core gRPC types
- `codes` & `status`: Untuk gRPC error handling

## Interceptor Implementations

### 1. LoggingInterceptor
```go
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		log.Printf("gRPC Request - Method: %s, Request: %+v", info.FullMethod, req)
		
		resp, err := handler(ctx, req)
		
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("gRPC Response - Method: %s, Duration: %v, Error: %v", info.FullMethod, duration, err)
		} else {
			log.Printf("gRPC Response - Method: %s, Duration: %v, Success", info.FullMethod, duration)
		}
		
		return resp, err
	}
}
```

**Line by Line Explanation:**

#### Function Declaration
```go
func LoggingInterceptor() grpc.UnaryServerInterceptor {
```
- Factory function yang return interceptor
- Pattern ini memungkinkan configuration (misal log level) di masa depan

#### Closure Return
```go
return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
```
- Return anonymous function yang implement interceptor signature
- Closure pattern untuk maintain state (jika diperlukan)

#### Start Timing
```go
start := time.Now()
```
- Record start time untuk duration calculation
- Performance monitoring purpose

#### Request Logging
```go
log.Printf("gRPC Request - Method: %s, Request: %+v", info.FullMethod, req)
```
- `info.FullMethod`: Full method name (e.g., "/user.UserService/CreateUser")
- `%+v`: Verbose formatting untuk struct fields
- Log semua incoming requests untuk debugging

#### Handler Execution
```go
resp, err := handler(ctx, req)
```
- Call next handler dalam chain
- Bisa interceptor lain atau actual business handler
- Pass through context dan request

#### Duration Calculation
```go
duration := time.Since(start)
```
- Calculate total request processing time
- Important untuk performance monitoring

#### Response Logging
```go
if err != nil {
	log.Printf("gRPC Response - Method: %s, Duration: %v, Error: %v", info.FullMethod, duration, err)
} else {
	log.Printf("gRPC Response - Method: %s, Duration: %v, Success", info.FullMethod, duration)
}
```
- Different log format untuk success vs error
- Include duration untuk performance analysis
- Error details untuk debugging

#### Return Response
```go
return resp, err
```
- Pass through response dan error ke client
- Interceptor tidak modify response, hanya observe

### 2. RecoveryInterceptor
```go
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered in gRPC handler %s: %v", info.FullMethod, r)
				err = status.Error(codes.Internal, "Internal server error")
			}
		}()
		
		return handler(ctx, req)
	}
}
```

**Line by Line Explanation:**

#### Named Return Values
```go
(resp interface{}, err error)
```
- Named return values untuk modify dalam defer function
- Memungkinkan defer function untuk set error value

#### Defer Recovery
```go
defer func() {
	if r := recover(); r != nil {
		log.Printf("Panic recovered in gRPC handler %s: %v", info.FullMethod, r)
		err = status.Error(codes.Internal, "Internal server error")
	}
}()
```
- `defer`: Execute setelah function return (atau panic)
- `recover()`: Catch panic dan return panic value
- Log panic untuk debugging
- Set `err` variable (named return) dengan proper gRPC error
- `codes.Internal`: HTTP 500 equivalent untuk gRPC
- Hide panic details dari client untuk security

#### Handler Execution
```go
return handler(ctx, req)
```
- Simple pass-through ke next handler
- Jika panic terjadi, defer function akan handle

### 3. ValidationInterceptor
```go
func ValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Add custom validation logic here if needed
		// For now, just pass through to handler
		return handler(ctx, req)
	}
}
```

**Penjelasan:**
- Placeholder untuk custom validation logic
- Bisa ditambah validation rules yang apply ke semua methods
- Contoh: rate limiting, authentication, request size validation
- Currently pass-through, tapi structure sudah ready

### 4. TimeoutInterceptor
```go
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		
		return handler(ctx, req)
	}
}
```

**Line by Line Explanation:**

#### Parameterized Factory
```go
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
```
- Factory function dengan parameter
- Memungkinkan configurable timeout per service

#### Context Timeout
```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```
- Create new context dengan timeout
- `cancel()` function untuk cleanup resources
- `defer cancel()`: Ensure cleanup bahkan jika panic
- Timeout akan trigger context cancellation

#### Handler with Timeout Context
```go
return handler(ctx, req)
```
- Pass modified context dengan timeout ke handler
- Handler bisa check `ctx.Done()` untuk timeout detection
- Database queries, external calls akan respect timeout

## Interceptor Chain Order

Ketika multiple interceptors digunakan:
```go
grpc.ChainUnaryInterceptor(
    interceptor.RecoveryInterceptor(),    // 1. Outermost - catch panics
    interceptor.LoggingInterceptor(),     // 2. Log requests/responses  
    interceptor.ValidationInterceptor(),  // 3. Validate requests
    interceptor.TimeoutInterceptor(),     // 4. Innermost - set timeout
)
```

**Execution Flow:**
1. **Recovery** wraps everything untuk catch panics
2. **Logging** logs request, then calls next
3. **Validation** validates request, then calls next  
4. **Timeout** sets timeout context, then calls next
5. **Actual Handler** executes business logic
6. Response flows back through chain in reverse order

## Best Practices Implemented

1. **Separation of Concerns**: Each interceptor has single responsibility
2. **Error Handling**: Proper gRPC error codes
3. **Performance Monitoring**: Request duration logging
4. **Security**: Hide internal errors dari clients
5. **Resource Management**: Proper context cancellation
6. **Configurability**: Parameterized interceptors
7. **Observability**: Comprehensive logging untuk debugging

## Usage Example
```go
// In server setup
server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        interceptor.RecoveryInterceptor(),
        interceptor.LoggingInterceptor(),
        interceptor.TimeoutInterceptor(30*time.Second),
    ),
)
```

Interceptors provide powerful way untuk add cross-cutting functionality tanpa pollute business logic!
