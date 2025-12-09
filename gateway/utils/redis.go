package utils

import (
	"context"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
	once   sync.Once
	config *Config
)

// InitRedis initializes the Redis client and returns an error if it fails
func InitRedis() error {
	// Load configuration if not already loaded
	if config == nil {
		config = LoadConfig()
	}

	var err error
	once.Do(func() {
		redisAddr := config.GetRedisAddr()

		Client = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})

		_, err = Client.Ping(Ctx).Result()
		if err != nil {
			log.Printf("Failed to connect to Redis: %v", err)
		} else {
			log.Printf("Connected to Redis at %s", redisAddr)
		}
	})

	return err
}

// CloseRedis closes the Redis connection
func CloseRedis() {
	if Client != nil {
		if err := Client.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("Redis connection closed")
		}
	}
}
