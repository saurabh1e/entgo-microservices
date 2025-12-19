package graph

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/pkg/authz"
	pkgcache "github.com/saurabh/entgo-microservices/pkg/cache"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/jwt"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

// cacheUserData loads user role/permissions and caches in Redis asynchronously
func (r *mutationResolver) cacheUserData(userEntity *ent.User) {
	go func() {
		bgCtx := authz.SetBypass(context.Background(), true)

		cacheData := &pkgcontext.CachedUserData{
			User: &pkgcontext.User{
				ID:       userEntity.ID,
				Username: userEntity.Username,
				Email:    userEntity.Email,
				Name:     userEntity.Name,
				IsActive: userEntity.IsActive,
			},
		}

		// Load user's role and permissions
		if role, err := userEntity.QueryRole().Only(bgCtx); err == nil {
			cacheData.Role = &pkgcontext.CachedRole{
				ID:          role.ID,
				Name:        role.Name,
				DisplayName: role.DisplayName,
				Priority:    role.Priority,
			}

			// Load permissions through role_permissions junction
			if rolePerms, err := role.QueryRolePermissions().WithPermission().All(bgCtx); err == nil {
				cacheData.Permissions = make([]pkgcontext.CachedPermission, 0, len(rolePerms))
				for _, rp := range rolePerms {
					if perm := rp.Edges.Permission; perm != nil {
						cacheData.Permissions = append(cacheData.Permissions, pkgcontext.CachedPermission{
							Name:      perm.Name,
							CanRead:   rp.CanRead,
							CanCreate: rp.CanCreate,
							CanUpdate: rp.CanUpdate,
							CanDelete: rp.CanDelete,
						})
					}
				}
			}
		}

		// Cache with 1 hour TTL (matching access token expiry)
		if err := pkgcache.SetUserInCache(bgCtx, r.redisClient, "auth", cacheData, 1*time.Hour); err != nil {
			logger.WithError(err).WithField("user_id", userEntity.ID).Warn("Failed to cache user data")
		} else {
			logger.WithFields(map[string]interface{}{
				"user_id":           userEntity.ID,
				"has_role":          cacheData.Role != nil,
				"permissions_count": len(cacheData.Permissions),
			}).Debug("User data cached")
		}
	}()
}

// hashPasswordIfNeeded checks if password is hashed, if not hashes and updates it
func (r *mutationResolver) hashPasswordIfNeeded(ctx context.Context, userEntity *ent.User) (string, error) {
	if strings.HasPrefix(userEntity.PasswordHash, "$2") { // bcrypt hash starts with $2
		return userEntity.PasswordHash, nil
	}

	hashedPassword, err := jwt.HashPassword(userEntity.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	if _, err = r.client.User.UpdateOneID(userEntity.ID).SetPasswordHash(hashedPassword).Save(ctx); err != nil {
		return "", fmt.Errorf("failed to update password hash: %w", err)
	}

	return hashedPassword, nil
}

// extractToken gets and cleans the authorization token from context
func extractToken(ctx context.Context) (string, error) {
	authHeader := ctx.Value("Authorization")
	if authHeader == nil {
		return "", fmt.Errorf("no authorization token provided")
	}

	tokenString, ok := authHeader.(string)
	if !ok {
		return "", fmt.Errorf("invalid authorization header")
	}

	return strings.TrimPrefix(tokenString, "Bearer "), nil
}
