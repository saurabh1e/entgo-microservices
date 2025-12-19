package hooks

import (
	"context"
	"fmt"
	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/hook"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func RoleCreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if roleMutation, ok := m.(*ent.RoleMutation); ok {
				// Hook executing for create

				// Set tenant ID from context for tenant-isolated entities
				if _, exists := roleMutation.TenantID(); !exists {
					tenantID, err := pkgcontext.GetUserTenantID(ctx)
					if err != nil {
						logger.WithError(err).WithFields(map[string]interface{}{
							"entity":    "Role",
							"operation": "create",
						}).Error("Failed to get tenant ID from context")
						return nil, err
					}
					roleMutation.SetTenantID(tenantID)
				}

				// Auto-generate code field from tenant_id and name (if name is available)
				if name, nameExists := roleMutation.Name(); nameExists {
					tenantID, tenantExists := roleMutation.TenantID()
					if !tenantExists {
						logger.WithFields(map[string]interface{}{
							"entity":    "Role",
							"operation": "create",
						}).Error("tenant_id is required for code generation")
						return nil, fmt.Errorf("tenant_id is required for code generation")
					}
					code := schema.GenerateCode(tenantID, name)
					roleMutation.SetCode(code)
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

func RoleBulkUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.RoleMutation); ok {
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

func RoleSingleUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if roleMutation, ok := m.(*ent.RoleMutation); ok {
				// Hook executing for single update
				_ = roleMutation

				// Note: code field is immutable, regeneration on update is not allowed

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

func RoleBulkDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.RoleMutation); ok {
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

func RoleSingleDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if roleMutation, ok := m.(*ent.RoleMutation); ok {
				// Hook executing for single delete
				_ = roleMutation

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

func RoleHooks() []ent.Hook {
	return []ent.Hook{
		// Execute RoleCreateHook only for Create operations
		hook.On(RoleCreateHook(), ent.OpCreate),
		// Execute RoleBulkUpdateHook only for bulk Update operations
		hook.On(RoleBulkUpdateHook(), ent.OpUpdate),
		// Execute RoleSingleUpdateHook only for single UpdateOne operations
		hook.On(RoleSingleUpdateHook(), ent.OpUpdateOne),
		// Execute RoleBulkDeleteHook only for bulk Delete operations
		hook.On(RoleBulkDeleteHook(), ent.OpDelete),
		// Execute RoleSingleDeleteHook only for single DeleteOne operations
		hook.On(RoleSingleDeleteHook(), ent.OpDeleteOne),
	}
}
