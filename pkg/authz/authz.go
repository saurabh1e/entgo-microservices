package authz

import (
	"context"

	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// Context key type for bypass authorization
type contextKey string

const (
	// BypassAuthzKey is the key for bypassing authorization checks
	BypassAuthzKey contextKey = "bypass_authz"
)

// SetBypass sets the bypass authorization flag in context
func SetBypass(ctx context.Context, bypass bool) context.Context {
	return context.WithValue(ctx, BypassAuthzKey, bypass)
}

// CheckBypass checks if authorization should be bypassed
// Returns "Allow" if bypass is enabled, "Skip" otherwise
func CheckBypass(ctx context.Context) (string, error) {
	if bypass, ok := ctx.Value(BypassAuthzKey).(bool); ok && bypass {
		return "Allow", nil
	}
	return "Skip", nil
}

// HasRole checks if user has the specified role
func HasRole(ctx context.Context, roleName string) bool {
	cachedData, err := pkgcontext.GetCachedUserData(ctx)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"required_role": roleName,
		}).Debug("Failed to get cached user data for role check")
		return false
	}

	if cachedData.Role == nil {
		logger.WithFields(map[string]interface{}{
			"required_role": roleName,
		}).Debug("User has no role assigned")
		return false
	}

	if cachedData.Role.Name == roleName {
		logger.WithFields(map[string]interface{}{
			"user_role":    cachedData.Role.Name,
			"matched_role": roleName,
		}).Debug("Role matched")
		return true
	}

	logger.WithFields(map[string]interface{}{
		"user_role":     cachedData.Role.Name,
		"required_role": roleName,
	}).Debug("Role does not match")
	return false
}

// HasAnyRole checks if user has any of the specified roles
func HasAnyRole(ctx context.Context, roleNames []string) bool {
	cachedData, err := pkgcontext.GetCachedUserData(ctx)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"required_roles": roleNames,
		}).Debug("Failed to get cached user data for role check")
		return false
	}

	if cachedData.Role == nil {
		logger.WithFields(map[string]interface{}{
			"required_roles": roleNames,
		}).Debug("User has no role assigned")
		return false
	}

	for _, roleName := range roleNames {
		if cachedData.Role.Name == roleName {
			logger.WithFields(map[string]interface{}{
				"user_role":    cachedData.Role.Name,
				"matched_role": roleName,
			}).Debug("Role matched")
			return true
		}
	}

	logger.WithFields(map[string]interface{}{
		"user_role":      cachedData.Role.Name,
		"required_roles": roleNames,
	}).Debug("No role match found")
	return false
}

// HasPermission checks if user has specific permission with action
// action should be one of: "can_read", "can_create", "can_update", "can_delete"
func HasPermission(ctx context.Context, permissionName string, action string) bool {
	cachedData, err := pkgcontext.GetCachedUserData(ctx)
	if err != nil {
		logger.WithError(err).WithFields(map[string]interface{}{
			"permission": permissionName,
			"action":     action,
		}).Debug("Failed to get cached user data for permission check")
		return false
	}

	// Find the permission
	for _, perm := range cachedData.Permissions {
		if perm.Name != permissionName {
			continue
		}

		logger.WithFields(map[string]interface{}{
			"permission": permissionName,
			"action":     action,
			"can_read":   perm.CanRead,
			"can_create": perm.CanCreate,
			"can_update": perm.CanUpdate,
			"can_delete": perm.CanDelete,
		}).Debug("Checking permission action")

		switch action {
		case "can_read":
			return perm.CanRead
		case "can_create":
			return perm.CanCreate
		case "can_update":
			return perm.CanUpdate
		case "can_delete":
			return perm.CanDelete
		default:
			logger.WithFields(map[string]interface{}{
				"permission": permissionName,
				"action":     action,
			}).Debug("Unknown action type")
			return false
		}
	}

	// Log with role name only if role is assigned
	logFields := map[string]interface{}{
		"permission": permissionName,
		"action":     action,
	}
	if cachedData.Role != nil {
		logFields["role"] = cachedData.Role.Name
	} else {
		logFields["role"] = "none"
	}

	logger.WithFields(logFields).Debug("Permission not found")
	return false
}
