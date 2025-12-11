package grpc

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GatewayClient provides a client connection to the gateway's gRPC proxy
// Microservices should use this to communicate with other services through the gateway
type GatewayClient struct {
	conn           *grpc.ClientConn
	gatewayAddr    string
	mu             sync.RWMutex
	useGateway     bool
	directClient   *ClientPool
	serviceClients *ServiceClients
	servicesOnce   sync.Once
	connOnce       sync.Once
	connErr        error
}

// NewGatewayClientLazy creates a new gateway client without establishing connection immediately
// Connection will be established on first use. This is useful when gateway starts after services.
// If GATEWAY_GRPC_URL is set, it will use the gateway proxy (lazy connection)
// Otherwise, it falls back to direct service-to-service connections
func NewGatewayClientLazy() *GatewayClient {
	gatewayAddr := os.Getenv("GATEWAY_GRPC_URL")
	useGateway := gatewayAddr != ""

	return &GatewayClient{
		gatewayAddr:  gatewayAddr,
		useGateway:   useGateway,
		directClient: NewClientPool(),
	}
}

// NewGatewayClient creates a new gateway client
// If GATEWAY_GRPC_URL is set, it uses the gateway proxy
// Otherwise, it falls back to direct service-to-service connections
func NewGatewayClient() (*GatewayClient, error) {
	gatewayAddr := os.Getenv("GATEWAY_GRPC_URL")
	useGateway := gatewayAddr != ""

	client := &GatewayClient{
		gatewayAddr:  gatewayAddr,
		useGateway:   useGateway,
		directClient: NewClientPool(),
	}

	if useGateway {
		// Connect to gateway proxy
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := grpc.NewClient(gatewayAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
				grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
			),
			grpc.WithChainUnaryInterceptor(
				ClientRetryInterceptor(3, 100*time.Millisecond),
				ClientLoggingInterceptor(),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to gateway proxy at %s: %w", gatewayAddr, err)
		}

		// Use context to avoid unused warning
		_ = ctx

		client.conn = conn
		fmt.Printf("✅ Connected to gRPC gateway proxy at %s\n", gatewayAddr)
	} else {
		fmt.Println("⚠️  GATEWAY_GRPC_URL not set, using direct service connections")
	}

	return client, nil
}

// GetConnection returns the gateway connection if using gateway proxy,
// or a direct connection to the specified service
func (c *GatewayClient) GetConnection(ctx context.Context, serviceAddr string) (*grpc.ClientConn, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.useGateway {
		// Return gateway connection - it will proxy to the appropriate service
		return c.conn, nil
	}

	// Fall back to direct connection
	return c.directClient.GetConnection(ctx, serviceAddr)
}

// ensureConnection establishes the gateway connection if not already connected (lazy connection)
func (c *GatewayClient) ensureConnection() error {
	if !c.useGateway {
		return nil // No connection needed for direct mode
	}

	c.connOnce.Do(func() {
		if c.conn != nil {
			return // Already connected
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := grpc.NewClient(c.gatewayAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                30 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			}),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
				grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
			),
			grpc.WithChainUnaryInterceptor(
				ClientRetryInterceptor(3, 100*time.Millisecond),
				ClientLoggingInterceptor(),
			),
		)
		if err != nil {
			c.connErr = fmt.Errorf("failed to connect to gateway proxy at %s: %w", c.gatewayAddr, err)
			return
		}

		// Use context to avoid unused warning
		_ = ctx

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()
		fmt.Printf("✅ Connected to gRPC gateway proxy at %s\n", c.gatewayAddr)
	})

	return c.connErr
}

// GetGatewayConnection returns the gateway proxy connection
// Returns error if gateway is not configured
func (c *GatewayClient) GetGatewayConnection() (*grpc.ClientConn, error) {
	if !c.useGateway {
		return nil, fmt.Errorf("gateway proxy not configured (GATEWAY_GRPC_URL not set)")
	}

	// Ensure connection is established
	if err := c.ensureConnection(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn, nil
}

// Services returns lazy-initialized service clients for all microservices
// Use this to access any microservice: gateway.Services().Auth().GetUserByID(...)
func (c *GatewayClient) Services() *ServiceClients {
	c.servicesOnce.Do(func() {
		c.serviceClients = NewServiceClients(c)
	})
	return c.serviceClients
}

// GetAutoClients is deprecated, use Services() instead
func (c *GatewayClient) GetAutoClients() *AutoClients {
	return c.Services()
}

// IsUsingGateway returns true if using gateway proxy, false for direct connections
func (c *GatewayClient) IsUsingGateway() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.useGateway
}

// Close closes the gateway connection
func (c *GatewayClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close gateway connection: %w", err)
		}
		c.conn = nil
	}

	if c.directClient != nil {
		return c.directClient.Close()
	}

	return nil
}

// Global gateway client instance (optional singleton pattern)
var (
	globalGatewayClient     *GatewayClient
	globalGatewayClientOnce sync.Once
	globalGatewayClientErr  error
)

// GetGlobalGatewayClient returns a singleton gateway client instance
func GetGlobalGatewayClient() (*GatewayClient, error) {
	globalGatewayClientOnce.Do(func() {
		globalGatewayClient, globalGatewayClientErr = NewGatewayClient()
	})
	return globalGatewayClient, globalGatewayClientErr
}
