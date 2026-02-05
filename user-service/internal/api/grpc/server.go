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

// Server represents gRPC server
type Server struct {
	server      *grpc.Server
	listener    net.Listener
	userHandler *handler.UserHandler
}

// NewServer creates new gRPC server
func NewServer(cfg *config.GRPCServerConfig, userHandler *handler.UserHandler) (*Server, error) {
	// Create listener
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", cfg.Port, err)
	}
	
	// Setup server options
	opts := []grpc.ServerOption{
		// Add interceptors
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggingInterceptor(),
			interceptor.ValidationInterceptor(),
			interceptor.TimeoutInterceptor(time.Duration(cfg.RequestTimeout)*time.Second),
		),
		
		// Keepalive settings
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		
		// Keepalive enforcement
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		
		// Max message size
		grpc.MaxRecvMsgSize(4 * 1024 * 1024), // 4MB
		grpc.MaxSendMsgSize(4 * 1024 * 1024), // 4MB
		
		// Max concurrent streams
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConnections)),
	}
	
	// Add TLS if enabled
	if cfg.EnableTLS {
		// TODO: Add TLS credentials
		// creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		// if err != nil {
		//     return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		// }
		// opts = append(opts, grpc.Creds(creds))
	}
	
	// Create gRPC server
	server := grpc.NewServer(opts...)
	
	// Register services
	pb.RegisterUserServiceServer(server, userHandler)
	
	// Enable reflection for development
	reflection.Register(server)
	
	return &Server{
		server:      server,
		listener:    lis,
		userHandler: userHandler,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	return s.server.Serve(s.listener)
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	s.server.GracefulStop()
}

// GetListener returns the server listener
func (s *Server) GetListener() net.Listener {
	return s.listener
}

// GetServer returns the gRPC server instance
func (s *Server) GetServer() *grpc.Server {
	return s.server
}
