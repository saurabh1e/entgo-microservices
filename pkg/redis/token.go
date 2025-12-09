package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenService manages token whitelist and blacklist using Redis with service namespacing
type TokenService struct {
	client      *redis.Client
	serviceName string
}

// NewTokenService creates a new Redis token service with service-specific namespacing
func NewTokenService(client *redis.Client, serviceName string) *TokenService {
	return &TokenService{
		client:      client,
		serviceName: serviceName,
	}
}

// buildKey creates a namespaced key with service prefix
func (r *TokenService) buildKey(resourceType, identifier string) string {
	return fmt.Sprintf("%s:%s:%s", r.serviceName, resourceType, identifier)
}

// AddToWhitelist adds a token to the whitelist with TTL
func (r *TokenService) AddToWhitelist(ctx context.Context, tokenID string, expiry time.Duration) error {
	key := r.buildKey("whitelist", tokenID)
	return r.client.Set(ctx, key, "valid", expiry).Err()
}

// IsWhitelisted checks if a token is in the whitelist
func (r *TokenService) IsWhitelisted(ctx context.Context, tokenID string) (bool, error) {
	key := r.buildKey("whitelist", tokenID)
	result := r.client.Exists(ctx, key)
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val() > 0, nil
}

// RemoveFromWhitelist removes a token from the whitelist
func (r *TokenService) RemoveFromWhitelist(ctx context.Context, tokenID string) error {
	key := r.buildKey("whitelist", tokenID)
	return r.client.Del(ctx, key).Err()
}

// AddToBlacklist adds a token to the blacklist with TTL
func (r *TokenService) AddToBlacklist(ctx context.Context, tokenID string, expiry time.Duration) error {
	key := r.buildKey("blacklist", tokenID)
	return r.client.Set(ctx, key, "revoked", expiry).Err()
}

// IsBlacklisted checks if a token is in the blacklist
func (r *TokenService) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := r.buildKey("blacklist", tokenID)
	result := r.client.Exists(ctx, key)
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val() > 0, nil
}

// RevokeToken moves a token from whitelist to blacklist
func (r *TokenService) RevokeToken(ctx context.Context, tokenID string, expiry time.Duration) error {
	// Add to blacklist
	if err := r.AddToBlacklist(ctx, tokenID, expiry); err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	// Remove from whitelist
	if err := r.RemoveFromWhitelist(ctx, tokenID); err != nil {
		return fmt.Errorf("failed to remove token from whitelist: %w", err)
	}

	return nil
}

// IsTokenValid checks if a token is valid (whitelisted and not blacklisted)
func (r *TokenService) IsTokenValid(ctx context.Context, tokenID string) (bool, error) {
	// Check if blacklisted first (faster lookup)
	blacklisted, err := r.IsBlacklisted(ctx, tokenID)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if blacklisted {
		return false, nil
	}

	// Check if whitelisted
	whitelisted, err := r.IsWhitelisted(ctx, tokenID)
	if err != nil {
		return false, fmt.Errorf("failed to check whitelist: %w", err)
	}

	return whitelisted, nil
}

// CleanupExpiredTokens removes expired tokens from both lists (optional cleanup)
func (r *TokenService) CleanupExpiredTokens(ctx context.Context) error {
	// Redis automatically handles TTL expiration, so this is mainly for monitoring
	// You can implement custom logic here if needed
	return nil
}
