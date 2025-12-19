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

func UserCreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if userMutation, ok := m.(*ent.UserMutation); ok {
				// Hook executing for create

				// Set tenant ID from context for tenant-isolated entities
				if _, exists := userMutation.TenantID(); !exists {
					tenantID, err := pkgcontext.GetUserTenantID(ctx)
					if err != nil {
						logger.WithError(err).WithFields(map[string]interface{}{
							"entity":    "User",
							"operation": "create",
						}).Error("Failed to get tenant ID from context")
						return nil, err
					}
					userMutation.SetTenantID(tenantID)
				}

				// Auto-generate code field from tenant_id and name (if name is available)
				if name, nameExists := userMutation.Name(); nameExists {
					tenantID, tenantExists := userMutation.TenantID()
					if !tenantExists {
						logger.WithFields(map[string]interface{}{
							"entity":    "User",
							"operation": "create",
						}).Error("tenant_id is required for code generation")
						return nil, fmt.Errorf("tenant_id is required for code generation")
					}
					code := schema.GenerateCode(tenantID, name)
					userMutation.SetCode(code)
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

func UserBulkUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.UserMutation); ok {
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

func UserSingleUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if userMutation, ok := m.(*ent.UserMutation); ok {
				// Hook executing for single update
				_ = userMutation

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

func UserBulkDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.UserMutation); ok {
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

func UserSingleDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if userMutation, ok := m.(*ent.UserMutation); ok {
				// Hook executing for single delete
				_ = userMutation

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

func UserHooks() []ent.Hook {
	return []ent.Hook{
		// Execute UserCreateHook only for Create operations
		hook.On(UserCreateHook(), ent.OpCreate),
		// Execute UserBulkUpdateHook only for bulk Update operations
		hook.On(UserBulkUpdateHook(), ent.OpUpdate),
		// Execute UserSingleUpdateHook only for single UpdateOne operations
		hook.On(UserSingleUpdateHook(), ent.OpUpdateOne),
		// Execute UserBulkDeleteHook only for bulk Delete operations
		hook.On(UserBulkDeleteHook(), ent.OpDelete),
		// Execute UserSingleDeleteHook only for single DeleteOne operations
		hook.On(UserSingleDeleteHook(), ent.OpDeleteOne),
	}
}
