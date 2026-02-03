package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"user-service/internal/config"
	"user-service/internal/infrastructure/grpclient"
	"user-service/proto/user"
	// Import lain akan ditambah setelah layer dibuat
)

func main() {
	// 1. Load Configuration
	cfg := config.Load()
	
	// 2. Initialize Database Connection
	db, err := config.NewPostgresConn(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// 3. Initialize RabbitMQ Connection
	rabbitmq, err := config.NewRabbitMQConn(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()
	
	// 4. Initialize gRPC Client Connections (untuk call service lain)
	grpcClients, err := grpclient.NewGRPCClients(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}
	defer grpcClients.Close()
	
	// 5. Initialize Repository Layer
	userRepo := repository.NewUserRepository(db)
	
	// 6. Initialize Service Layer
	userService := services.NewUserService(userRepo, rabbitmq, grpcClients)
	
	// 7. Initialize gRPC Handler
	userHandler := handler.NewUserHandler(userService)
	
	// 8. Setup gRPC Server
	grpcConfig := config.NewGRPCServerConfig(cfg)
	lis, err := net.Listen("tcp", ":"+grpcConfig.Port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	server := grpc.NewServer()
	user.RegisterUserServiceServer(server, userHandler)
	
	log.Printf("gRPC server listening on port %s", grpcConfig.Port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
