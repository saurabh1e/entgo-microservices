package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientPool manages gRPC client connections with connection pooling
type ClientPool struct {
	connections map[string]*grpc.ClientConn
	mu          sync.RWMutex
}

// NewClientPool creates a new client connection pool
func NewClientPool() *ClientPool {
	return &ClientPool{
		connections: make(map[string]*grpc.ClientConn),
	}
}

// GetConnection returns an existing connection or creates a new one
func (p *ClientPool) GetConnection(ctx context.Context, serviceAddr string) (*grpc.ClientConn, error) {
	// Check if connection exists
	p.mu.RLock()
	if conn, exists := p.connections[serviceAddr]; exists {
		p.mu.RUnlock()
		return conn, nil
	}
	p.mu.RUnlock()

	// Create new connection with write lock
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists := p.connections[serviceAddr]; exists {
		return conn, nil
	}

	// Create new connection with retry and logging interceptors
	conn, err := grpc.DialContext(ctx, serviceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithChainUnaryInterceptor(
			ClientRetryInterceptor(3, 100*time.Millisecond),
			ClientLoggingInterceptor(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", serviceAddr, err)
	}

	p.connections[serviceAddr] = conn
	return conn, nil
}

// Close closes all connections in the pool
func (p *ClientPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for addr, conn := range p.connections {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection to %s: %w", addr, err)
		}
	}
	p.connections = make(map[string]*grpc.ClientConn)
	return nil
}
