package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/saurabh/entgo-microservices/pkg/jwt"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtService *jwt.Service
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtService *jwt.Service) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// Middleware extracts and validates JWT tokens from requests
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Remove "Bearer " prefix if present
			if strings.HasPrefix(authHeader, "Bearer ") {
				authHeader = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// Add token to context for resolvers to use
			ctx := context.WithValue(r.Context(), "Authorization", authHeader)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuth middleware that requires valid authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		token := authHeader
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		// Validate token
		claims, err := m.jwtService.ValidateToken(r.Context(), token)
		if err != nil {
			logger.WithError(err).Debug("Token validation failed")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "user_claims", claims)
		ctx = context.WithValue(ctx, "Authorization", token)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
