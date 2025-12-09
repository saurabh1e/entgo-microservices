package graph

import (
	"github.com/saurabh/entgo-microservices/test-service/internal/ent"

	"github.com/redis/go-redis/v9"
	pkggrpc "github.com/saurabh/entgo-microservices/pkg/grpc"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client      *ent.Client
	jwtService  *jwt.Service
	redisClient *redis.Client
	grpcPool    *pkggrpc.ClientPool
	userClient  *pkggrpc.UserClient
}

func NewResolver(client *ent.Client, jwtService *jwt.Service, redisClient *redis.Client, grpcPool *pkggrpc.ClientPool, userClient *pkggrpc.UserClient) *Resolver {
	return &Resolver{
		client:      client,
		jwtService:  jwtService,
		redisClient: redisClient,
		grpcPool:    grpcPool,
		userClient:  userClient,
	}
}
