package graph

import (
	"github.com/saurabh/entgo-microservices/auth/internal/ent"

	"github.com/redis/go-redis/v9"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	client      *ent.Client
	jwtService  *jwt.Service
	redisClient *redis.Client
}

func NewResolver(client *ent.Client, jwtService *jwt.Service, redisClient *redis.Client) *Resolver {
	return &Resolver{
		client:      client,
		jwtService:  jwtService,
		redisClient: redisClient,
	}
}
