package hooks

import (
	"context"

	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/hook"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func UserCreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.UserMutation); ok {
				logger.WithFields(map[string]interface{}{
					"entity":    "User",
					"operation": "create",
				}).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-create logic here
				if err == nil {
					logger.WithFields(map[string]interface{}{
						"entity":    "User",
						"operation": "create",
					}).Debug("Hook completed successfully")
				}

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
				logger.WithFields(map[string]interface{}{
					"entity":    "User",
					"operation": "bulk_update",
				}).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-update logic here
				if err == nil {
					logger.WithFields(map[string]interface{}{
						"entity":    "User",
						"operation": "bulk_update",
					}).Debug("Hook completed successfully")
				}

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
				fields := map[string]interface{}{
					"entity":    "User",
					"operation": "update",
				}

				// Get the specific ID being updated for UpdateOne operations
				if id, exists := userMutation.ID(); exists {
					fields["id"] = id
				}

				logger.WithFields(fields).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-single-update logic here
				if err == nil {
					logger.WithFields(fields).Debug("Hook completed successfully")
				}

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
				logger.WithFields(map[string]interface{}{
					"entity":    "User",
					"operation": "bulk_delete",
				}).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-delete logic here
				if err == nil {
					logger.WithFields(map[string]interface{}{
						"entity":    "User",
						"operation": "bulk_delete",
					}).Debug("Hook completed successfully")
				}

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
				fields := map[string]interface{}{
					"entity":    "User",
					"operation": "delete",
				}

				// Get the specific ID being deleted for DeleteOne operations
				if id, exists := userMutation.ID(); exists {
					fields["id"] = id
				}

				logger.WithFields(fields).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-single-delete logic here
				if err == nil {
					logger.WithFields(fields).Debug("Hook completed successfully")
				}

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
