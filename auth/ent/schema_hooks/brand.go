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

func BrandCreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if brandMutation, ok := m.(*ent.BrandMutation); ok {
				// Hook executing for create

				// Set tenant ID from context for tenant-isolated entities
				if _, exists := brandMutation.TenantID(); !exists {
					tenantID, err := pkgcontext.GetUserTenantID(ctx)
					if err != nil {
						logger.WithError(err).WithFields(map[string]interface{}{
							"entity":    "Brand",
							"operation": "create",
						}).Error("Failed to get tenant ID from context")
						return nil, err
					}
					brandMutation.SetTenantID(tenantID)
				}

				// Auto-generate code field from tenant_id and name
				if name, nameExists := brandMutation.Name(); nameExists {
					if tenantID, tenantExists := brandMutation.TenantID(); tenantExists {
						code := schema.GenerateCode(tenantID, name)
						brandMutation.SetCode(code)
					} else {
						logger.WithFields(map[string]interface{}{
							"entity":    "Brand",
							"operation": "create",
						}).Error("tenant_id is required for code generation")
						return nil, fmt.Errorf("tenant_id is required for code generation")
					}
				} else {
					logger.WithFields(map[string]interface{}{
						"entity":    "Brand",
						"operation": "create",
					}).Error("name is required for code generation")
					return nil, fmt.Errorf("name is required for code generation")
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

func BrandBulkUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.BrandMutation); ok {
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

func BrandSingleUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if brandMutation, ok := m.(*ent.BrandMutation); ok {
				// Hook executing for single update
				_ = brandMutation

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

func BrandBulkDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.BrandMutation); ok {
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

func BrandSingleDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if brandMutation, ok := m.(*ent.BrandMutation); ok {
				// Hook executing for single delete
				_ = brandMutation

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

func BrandHooks() []ent.Hook {
	return []ent.Hook{
		// Execute BrandCreateHook only for Create operations
		hook.On(BrandCreateHook(), ent.OpCreate),
		// Execute BrandBulkUpdateHook only for bulk Update operations
		hook.On(BrandBulkUpdateHook(), ent.OpUpdate),
		// Execute BrandSingleUpdateHook only for single UpdateOne operations
		hook.On(BrandSingleUpdateHook(), ent.OpUpdateOne),
		// Execute BrandBulkDeleteHook only for bulk Delete operations
		hook.On(BrandBulkDeleteHook(), ent.OpDelete),
		// Execute BrandSingleDeleteHook only for single DeleteOne operations
		hook.On(BrandSingleDeleteHook(), ent.OpDeleteOne),
	}
}
