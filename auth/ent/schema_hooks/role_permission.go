package hooks

import (
	"context"
	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/hook"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func RolePermissionCreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if role_permissionMutation, ok := m.(*ent.RolePermissionMutation); ok {
				// Hook executing for create

				// Set tenant ID from context for tenant-isolated entities
				if _, exists := role_permissionMutation.TenantID(); !exists {
					tenantID, err := pkgcontext.GetUserTenantID(ctx)
					if err != nil {
						logger.WithError(err).WithFields(map[string]interface{}{
							"entity":    "RolePermission",
							"operation": "create",
						}).Error("Failed to get tenant ID from context")
						return nil, err
					}
					role_permissionMutation.SetTenantID(tenantID)
				}

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-create logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func RolePermissionBulkUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.RolePermissionMutation); ok {
				// Hook executing for bulk update

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-update logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func RolePermissionSingleUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if role_permissionMutation, ok := m.(*ent.RolePermissionMutation); ok {
				// Hook executing for single update
				_ = role_permissionMutation

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-single-update logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func RolePermissionBulkDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.RolePermissionMutation); ok {
				// Hook executing for bulk delete

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-delete logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func RolePermissionSingleDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if role_permissionMutation, ok := m.(*ent.RolePermissionMutation); ok {
				// Hook executing for single delete
				_ = role_permissionMutation

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-single-delete logic here (no verbose logs by default)
				_ = result
				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func RolePermissionHooks() []ent.Hook {
	return []ent.Hook{
		// Execute RolePermissionCreateHook only for Create operations
		hook.On(RolePermissionCreateHook(), ent.OpCreate),
		// Execute RolePermissionBulkUpdateHook only for bulk Update operations
		hook.On(RolePermissionBulkUpdateHook(), ent.OpUpdate),
		// Execute RolePermissionSingleUpdateHook only for single UpdateOne operations
		hook.On(RolePermissionSingleUpdateHook(), ent.OpUpdateOne),
		// Execute RolePermissionBulkDeleteHook only for bulk Delete operations
		hook.On(RolePermissionBulkDeleteHook(), ent.OpDelete),
		// Execute RolePermissionSingleDeleteHook only for single DeleteOne operations
		hook.On(RolePermissionSingleDeleteHook(), ent.OpDeleteOne),
	}
}
