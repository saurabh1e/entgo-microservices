package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis connection configuration
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewClient creates a new Redis client with the provided configuration
func NewClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// KeyBuilder helps construct namespaced keys for Redis
type KeyBuilder struct {
	service string
}

// NewKeyBuilder creates a new key builder for a specific service
func NewKeyBuilder(service string) *KeyBuilder {
	return &KeyBuilder{service: service}
}

// Build constructs a namespaced key with the pattern: {service}:{resourceType}:{identifier}
func (kb *KeyBuilder) Build(resourceType, identifier string) string {
	return fmt.Sprintf("%s:%s:%s", kb.service, resourceType, identifier)
}
