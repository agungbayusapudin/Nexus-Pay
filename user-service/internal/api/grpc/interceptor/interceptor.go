package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs all gRPC requests and responses
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

// RecoveryInterceptor recovers from panics and returns proper gRPC error
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

// ValidationInterceptor validates request before processing
func ValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Add custom validation logic here if needed
		// For now, just pass through to handler
		return handler(ctx, req)
	}
}

// TimeoutInterceptor adds timeout to requests
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		
		return handler(ctx, req)
	}
}
