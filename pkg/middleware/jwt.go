package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// JWTAuthMiddleware validates JWT tokens and adds user to context from Redis cache
type JWTAuthMiddleware struct {
	jwtService  *jwt.Service
	redisClient *redis.Client
	serviceName string
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(jwtService *jwt.Service, redisClient *redis.Client, serviceName string) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		jwtService:  jwtService,
		redisClient: redisClient,
		serviceName: serviceName,
	}
}

// buildUserKey creates Redis key for user data
func (m *JWTAuthMiddleware) buildUserKey(userID int) string {
	return fmt.Sprintf("%s:user:%d", m.serviceName, userID)
}

// GetUserFromCache retrieves user data from Redis cache
func (m *JWTAuthMiddleware) GetUserFromCache(ctx context.Context, userID int) (*pkgcontext.CachedUserData, error) {
	key := m.buildUserKey(userID)

	data, err := m.redisClient.Get(ctx, key).Result()
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

// Middleware validates JWT tokens and adds user to context from Redis cache
// Authentication is optional - if no token provided or invalid, request continues without user
// Use @auth directive in GraphQL schema to require authentication for specific queries/mutations
func (m *JWTAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Remove "Bearer " prefix if present
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := m.jwtService.ValidateToken(r.Context(), token)
		if err != nil {
			logger.WithError(err).Debug("Token validation failed")
			next.ServeHTTP(w, r)
			return
		}

		// Ensure it's an access token
		if claims.TokenType != "access" {
			logger.Debug("Invalid token type - not an access token")
			next.ServeHTTP(w, r)
			return
		}

		// Get user from Redis cache
		cachedData, err := m.GetUserFromCache(r.Context(), claims.UserID)
		if err != nil {
			logger.WithError(err).WithField("user_id", claims.UserID).Debug("User not found in cache")
			next.ServeHTTP(w, r)
			return
		}

		if cachedData == nil || cachedData.User == nil {
			logger.WithField("user_id", claims.UserID).Debug("User data is nil in cache")
			next.ServeHTTP(w, r)
			return
		}

		user := cachedData.User

		// Check if user is active
		if !user.IsActive {
			logger.WithField("user_id", user.ID).Debug("User account is deactivated")
			next.ServeHTTP(w, r)
			return
		}

		logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}).Debug("User authenticated from cache")

		// Add user to context
		ctx := pkgcontext.SetUser(r.Context(), user)
		ctx = pkgcontext.SetClaims(ctx, claims)
		ctx = pkgcontext.SetToken(ctx, token)

		// Store cached data in context for authorization checks
		ctx = pkgcontext.SetCachedUserData(ctx, cachedData)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
