package graph

import (
	"sync"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"

	"github.com/redis/go-redis/v9"
	pkggrpc "github.com/saurabh/entgo-microservices/pkg/grpc"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client        *ent.Client
	jwtService    *jwt.Service
	redisClient   *redis.Client
	gatewayClient *pkggrpc.GatewayClient
	gatewayOnce   sync.Once
	gatewayErr    error
}

func NewResolver(client *ent.Client, jwtService *jwt.Service, redisClient *redis.Client, gatewayClient *pkggrpc.GatewayClient) *Resolver {
	return &Resolver{
		client:        client,
		jwtService:    jwtService,
		redisClient:   redisClient,
		gatewayClient: gatewayClient,
	}
}

// GetGatewayClient returns the gateway client, initializing it on first call (lazy loading)
// This is needed because gateway starts after microservices
func (r *Resolver) GetGatewayClient() (*pkggrpc.GatewayClient, error) {
	r.gatewayOnce.Do(func() {
		if r.gatewayClient == nil {
			r.gatewayClient, r.gatewayErr = pkggrpc.NewGatewayClient()
		}
	})
	return r.gatewayClient, r.gatewayErr
}
