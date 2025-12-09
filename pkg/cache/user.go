package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
)

// BuildUserCacheKey creates a Redis key for user data
func BuildUserCacheKey(serviceName string, userID int) string {
	return fmt.Sprintf("%s:user:%d", serviceName, userID)
}

// SetUserInCache stores user data with role and permissions in Redis
func SetUserInCache(ctx context.Context, client *redis.Client, serviceName string, userData *pkgcontext.CachedUserData, ttl time.Duration) error {
	if userData == nil || userData.User == nil {
		return fmt.Errorf("user data cannot be nil")
	}

	key := BuildUserCacheKey(serviceName, userData.User.ID)
	userData.CachedAt = time.Now()

	data, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	if err := client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to cache user data: %w", err)
	}

	return nil
}

// GetUserFromCache retrieves user data from Redis
func GetUserFromCache(ctx context.Context, client *redis.Client, serviceName string, userID int) (*pkgcontext.CachedUserData, error) {
	key := BuildUserCacheKey(serviceName, userID)

	data, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found in cache")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user from cache: %w", err)
	}

	var userData pkgcontext.CachedUserData
	if err := json.Unmarshal([]byte(data), &userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &userData, nil
}

// InvalidateUserCache removes user data from Redis cache
func InvalidateUserCache(ctx context.Context, client *redis.Client, serviceName string, userID int) error {
	key := BuildUserCacheKey(serviceName, userID)
	return client.Del(ctx, key).Err()
}
