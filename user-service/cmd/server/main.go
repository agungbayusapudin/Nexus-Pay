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

func main() {
	// 1. Load Configuration
	cfg := config.Load()
	log.Printf("Starting user-service on port %s", cfg.GRPCPort)
	
	// 2. Initialize Database Connection
	db, err := config.NewPostgresConn(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")
	
	// 3. Initialize RabbitMQ Connection
	rabbitmq, err := config.NewRabbitMQConn(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()
	log.Println("RabbitMQ connected successfully")
	
	// 4. Initialize gRPC Client Connections (untuk call service lain)
	grpcClients, err := grpclient.NewGRPCClients(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}
	defer grpcClients.Close()
	log.Println("gRPC clients initialized successfully")
	
	// 5. Initialize Repository Layer
	// userRepo := repository.NewUserRepository(db)
	
	// 6. Initialize Service Layer
	// userService := services.NewUserService(userRepo, rabbitmq, grpcClients)
	
	// 7. Initialize gRPC Handler
	// userHandler := handler.NewUserHandler(userService)
	
	// Temporary: Create empty handler for now
	userHandler := handler.NewUserHandler(nil)
	
	// 8. Setup gRPC Server
	grpcConfig := config.NewGRPCServerConfig(cfg)
	server, err := grpcserver.NewServer(grpcConfig, userHandler)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}
	
	// 9. Start server in goroutine
	go func() {
		log.Printf("gRPC server listening on port %s", cfg.GRPCPort)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	
	// 10. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	// Graceful shutdown
	server.Stop()
	log.Println("Server stopped gracefully")
}
