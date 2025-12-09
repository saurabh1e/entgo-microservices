package graphql

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// AuthDirective is a GraphQL directive that checks if user is authenticated
// Use it in your schema: directive @auth on FIELD_DEFINITION
// Apply it to fields: me: User @auth
func AuthDirective(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Check if user exists in context
	user, ok := pkgcontext.GetUser(ctx)
	if !ok || user == nil {
		logger.Debug("Auth directive: no user in context")
		return nil, fmt.Errorf("unauthorized: authentication required")
	}

	// Check if user is active
	if !user.IsActive {
		logger.WithField("user_id", user.ID).Debug("Auth directive: user is inactive")
		return nil, fmt.Errorf("unauthorized: account is inactive")
	}

	logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	}).Debug("Auth directive: user authenticated")

	// User is authenticated, proceed with the resolver
	return next(ctx)
}

// HasRoleDirective checks that the authenticated user has the required role.
// Schema usage: directive @hasRole(role: String!) on FIELD_DEFINITION
func HasRoleDirective(ctx context.Context, obj interface{}, next graphql.Resolver, role string) (interface{}, error) {
	// Require authentication first
	user, ok := pkgcontext.GetUser(ctx)
	if !ok || user == nil {
		logger.Debug("HasRole directive: no user in context")
		return nil, fmt.Errorf("unauthorized: authentication required")
	}

	// Try to get cached user data (includes role & permissions)
	cached, err := pkgcontext.GetCachedUserData(ctx)
	if err != nil || cached == nil || cached.Role == nil {
		logger.WithField("user_id", user.ID).Debug("HasRole directive: cached role not available in context")
		return nil, fmt.Errorf("forbidden: role information not available")
	}

	if cached.Role.Name != role {
		logger.WithFields(map[string]interface{}{
			"user_id":   user.ID,
			"required":  role,
			"user_role": cached.Role.Name,
		}).Debug("HasRole directive: role mismatch")
		return nil, fmt.Errorf("forbidden: requires role '%s'", role)
	}

	logger.WithFields(map[string]interface{}{
		"user_id":   user.ID,
		"required":  role,
		"user_role": cached.Role.Name,
	}).Debug("HasRole directive: role check passed")

	return next(ctx)
}

// HasPermissionDirective checks that the authenticated user has a specific permission.
// Schema usage: directive @hasPermission(permission: String!) on FIELD_DEFINITION
func HasPermissionDirective(ctx context.Context, obj interface{}, next graphql.Resolver, permission string) (interface{}, error) {
	// Require authentication first
	user, ok := pkgcontext.GetUser(ctx)
	if !ok || user == nil {
		logger.Debug("HasPermission directive: no user in context")
		return nil, fmt.Errorf("unauthorized: authentication required")
	}

	// Get cached permissions
	cached, err := pkgcontext.GetCachedUserData(ctx)
	if err != nil || cached == nil {
		logger.WithField("user_id", user.ID).Debug("HasPermission directive: cached permissions not available in context")
		return nil, fmt.Errorf("forbidden: permission information not available")
	}

	// Look for matching permission name
	for _, p := range cached.Permissions {
		if p.Name == permission {
			logger.WithFields(map[string]interface{}{
				"user_id":    user.ID,
				"permission": permission,
			}).Debug("HasPermission directive: permission check passed")
			return next(ctx)
		}
	}

	logger.WithFields(map[string]interface{}{
		"user_id":    user.ID,
		"permission": permission,
	}).Debug("HasPermission directive: missing permission")

	return nil, fmt.Errorf("forbidden: requires permission '%s'", permission)
}
