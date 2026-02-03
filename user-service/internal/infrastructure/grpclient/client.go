package grpclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"user-service/internal/config"
)

// GRPCClients holds connections to other services
type GRPCClients struct {
	WalletConn      *grpc.ClientConn
	TransactionConn *grpc.ClientConn
	// WalletClient    wallet.WalletServiceClient    // Uncomment when proto ready
	// TransactionClient transaction.TransactionServiceClient
}

// NewGRPCClients creates connections to other microservices
func NewGRPCClients(cfg *config.Config) (*GRPCClients, error) {
	clients := &GRPCClients{}
	
	// 1. Connect to Wallet Service
	walletConn, err := createGRPCConnection(cfg.WalletServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to wallet service: %w", err)
	}
	clients.WalletConn = walletConn
	// clients.WalletClient = wallet.NewWalletServiceClient(walletConn)
	
	// 2. Connect to Transaction Service
	transactionConn, err := createGRPCConnection(cfg.TransactionServiceURL)
	if err != nil {
		clients.Close() // Cleanup previous connections
		return nil, fmt.Errorf("failed to connect to transaction service: %w", err)
	}
	clients.TransactionConn = transactionConn
	// clients.TransactionClient = transaction.NewTransactionServiceClient(transactionConn)
	
	return clients, nil
}

// createGRPCConnection creates a single gRPC connection
func createGRPCConnection(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Wait for connection to be ready
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", address, err)
	}
	
	return conn, nil
}

// Close closes all gRPC connections
func (g *GRPCClients) Close() error {
	var lastErr error
	
	if g.WalletConn != nil {
		if err := g.WalletConn.Close(); err != nil {
			lastErr = err
		}
	}
	
	if g.TransactionConn != nil {
		if err := g.TransactionConn.Close(); err != nil {
			lastErr = err
		}
	}
	
	return lastErr
}

// Health check methods
func (g *GRPCClients) IsWalletServiceHealthy() bool {
	if g.WalletConn == nil {
		return false
	}
	return g.WalletConn.GetState().String() == "READY"
}

func (g *GRPCClients) IsTransactionServiceHealthy() bool {
	if g.TransactionConn == nil {
		return false
	}
	return g.TransactionConn.GetState().String() == "READY"
}
