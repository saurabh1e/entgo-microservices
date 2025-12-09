package database

import (
	"context"
	"fmt"
	"time"

	"github.com/saurabh/entgo-microservices/test-service/config"

	"github.com/redis/go-redis/v9"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// RedisClient wraps the Redis client with additional functionality.
// It provides a unified interface for Redis operations with built-in
// connection management and health checking capabilities.
type RedisClient struct {
	*redis.Client
	closed bool
}

// NewRedisClient creates and initializes a new Redis client.
// It configures the connection, tests connectivity, and returns a ready-to-use client.
// Returns a RedisClient instance or an error if the connection fails.
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	logger.WithFields(map[string]interface{}{
		"host": cfg.Redis.Host,
		"port": cfg.Redis.Port,
		"db":   cfg.Redis.DB,
	}).Debug("Creating new Redis client")

	// Create Redis client with configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		// Close the client on connection failure to prevent resource leak
		if closeErr := rdb.Close(); closeErr != nil {
			logger.WithError(closeErr).Warn("Failed to close Redis client after connection failure")
		}
		logger.WithError(err).Error("Failed to connect to Redis")
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"host": cfg.Redis.Host,
		"port": cfg.Redis.Port,
		"db":   cfg.Redis.DB,
	}).Info("Redis client connected successfully")

	return &RedisClient{
		Client: rdb,
		closed: false,
	}, nil
}

// Close gracefully closes the Redis connection.
// It ensures that the connection is properly released and can be called multiple times safely.
// Returns an error if the close operation fails.
func (r *RedisClient) Close() error {
	if r.closed {
		logger.Debug("Redis connection already closed")
		return nil
	}

	logger.Info("Closing Redis connection")
	r.closed = true

	if err := r.Client.Close(); err != nil {
		logger.WithError(err).Error("Failed to close Redis connection")
		return fmt.Errorf("failed to close redis connection: %w", err)
	}

	logger.Info("Redis connection closed successfully")
	return nil
}

// HealthCheck performs a health check on the Redis connection.
// It verifies that the connection is still active and responsive.
// Returns an error if the connection is not healthy or if the client is closed.
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	if r.closed {
		return fmt.Errorf("redis client is closed")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if _, err := r.Ping(ctx).Result(); err != nil {
		logger.WithError(err).Warn("Redis health check failed")
		return fmt.Errorf("redis health check failed: %w", err)
	}

	logger.Debug("Redis health check passed")
	return nil
}
