package context

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Context key type to avoid collisions
type contextKey string

const (
	// UserCtxKey is the key for storing user objects in context
	UserCtxKey contextKey = "user"
	// ClaimsCtxKey is the key for storing JWT claims in context
	ClaimsCtxKey contextKey = "claims"
	// TokenCtxKey is the key for storing the raw token in context
	TokenCtxKey contextKey = "token"
	// CachedUserDataCtxKey is the key for storing complete cached user data (including role and permissions)
	CachedUserDataCtxKey contextKey = "cached_user_data"
)

// User represents a user in the context with common fields
// Services can embed this struct and add their own fields
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	IsActive bool   `json:"is_active"`
	TenantID int    `json:"tenant_id"`
}

// CachedUserData represents the complete user data with role and permissions
// This is stored in Redis and context by authentication middleware
type CachedUserData struct {
	User        *User              `json:"user"`
	Role        *CachedRole        `json:"role,omitempty"`
	Permissions []CachedPermission `json:"permissions,omitempty"`
	CachedAt    time.Time          `json:"cached_at"`
}

// CachedRole represents role information in cache
type CachedRole struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Priority    int    `json:"priority"`
}

// CachedPermission represents permission information in cache
type CachedPermission struct {
	Name      string `json:"name"`
	CanRead   bool   `json:"can_read"`
	CanCreate bool   `json:"can_create"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

// ContextData holds all authentication-related context data
type ContextData struct {
	User   *User
	Claims interface{} // JWT claims (type varies by service)
	Token  string      // Raw JWT token
}

// SetUser sets a User object in the context
func SetUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, UserCtxKey, user)
}

// GetUser retrieves a User object from the context
// Returns the user and a boolean indicating if it was found
func GetUser(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserCtxKey).(*User)
	return user, ok
}

// GetUserOrError retrieves a User from context or returns an error
func GetUserOrError(ctx context.Context) (*User, error) {
	user, ok := GetUser(ctx)
	if !ok || user == nil {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}

// GetUserTenantID retrieves the tenant ID from the user in context
func GetUserTenantID(ctx context.Context) (int, error) {
	user, ok := GetUser(ctx)
	if !ok || user == nil {
		return 0, errors.New("user not found in context")
	}
	if user.TenantID == 0 {
		return 0, errors.New("tenant ID not set for user")
	}
	return user.TenantID, nil
}

// RequireTenantID ensures a valid tenant ID exists in context, returns error if not
func RequireTenantID(ctx context.Context) error {
	_, err := GetUserTenantID(ctx)
	if err != nil {
		return fmt.Errorf("tenant context required: %w", err)
	}
	return nil
}

// SetClaims sets JWT claims in the context
func SetClaims(ctx context.Context, claims interface{}) context.Context {
	return context.WithValue(ctx, ClaimsCtxKey, claims)
}

// GetClaims retrieves JWT claims from the context
func GetClaims(ctx context.Context) (interface{}, bool) {
	claims := ctx.Value(ClaimsCtxKey)
	return claims, claims != nil
}

// SetToken sets the raw JWT token in the context
func SetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, TokenCtxKey, token)
}

// GetToken retrieves the raw JWT token from the context
func GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(TokenCtxKey).(string)
	return token, ok
}

// SetCachedUserData sets cached user data in context
func SetCachedUserData(ctx context.Context, data *CachedUserData) context.Context {
	return context.WithValue(ctx, CachedUserDataCtxKey, data)
}

// GetCachedUserData retrieves cached user data from context
func GetCachedUserData(ctx context.Context) (*CachedUserData, error) {
	data, ok := ctx.Value(CachedUserDataCtxKey).(*CachedUserData)
	if !ok || data == nil {
		return nil, fmt.Errorf("cached user data not found in context")
	}
	return data, nil
}

// SetContextData sets all authentication data in the context at once
func SetContextData(ctx context.Context, data *ContextData) context.Context {
	ctx = SetUser(ctx, data.User)
	ctx = SetClaims(ctx, data.Claims)
	ctx = SetToken(ctx, data.Token)
	return ctx
}

// GetContextData retrieves all authentication data from the context
func GetContextData(ctx context.Context) *ContextData {
	data := &ContextData{}

	if user, ok := GetUser(ctx); ok {
		data.User = user
	}

	if claims, ok := GetClaims(ctx); ok {
		data.Claims = claims
	}

	if token, ok := GetToken(ctx); ok {
		data.Token = token
	}

	return data
}
