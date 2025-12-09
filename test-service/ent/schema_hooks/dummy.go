package hooks

import (
	"context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
	"github.com/saurabh/entgo-microservices/test-service/internal/ent"
	"github.com/saurabh/entgo-microservices/test-service/internal/ent/hook"
)

func DummyCreateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.DummyMutation); ok {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"operation": "create",
				}).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-create logic here
				if err == nil {
					logger.WithFields(map[string]interface{}{
						"entity":    "Dummy",
						"operation": "create",
					}).Debug("Hook completed successfully")
				}

				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func DummyBulkUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.DummyMutation); ok {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"operation": "bulk_update",
				}).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-update logic here
				if err == nil {
					logger.WithFields(map[string]interface{}{
						"entity":    "Dummy",
						"operation": "bulk_update",
					}).Debug("Hook completed successfully")
				}

				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func DummySingleUpdateHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if dummyMutation, ok := m.(*ent.DummyMutation); ok {
				fields := map[string]interface{}{
					"entity":    "Dummy",
					"operation": "update",
				}

				// Get the specific ID being updated for UpdateOne operations
				if id, exists := dummyMutation.ID(); exists {
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

func DummyBulkDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if _, ok := m.(*ent.DummyMutation); ok {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"operation": "bulk_delete",
				}).Debug("Hook executing")

				// Call the next mutator
				result, err := next.Mutate(ctx, m)

				// Post-bulk-delete logic here
				if err == nil {
					logger.WithFields(map[string]interface{}{
						"entity":    "Dummy",
						"operation": "bulk_delete",
					}).Debug("Hook completed successfully")
				}

				return result, err
			}
			return next.Mutate(ctx, m)
		})
	}
}

func DummySingleDeleteHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if dummyMutation, ok := m.(*ent.DummyMutation); ok {
				fields := map[string]interface{}{
					"entity":    "Dummy",
					"operation": "delete",
				}

				// Get the specific ID being deleted for DeleteOne operations
				if id, exists := dummyMutation.ID(); exists {
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

func DummyHooks() []ent.Hook {
	return []ent.Hook{
		// Execute DummyCreateHook only for Create operations
		hook.On(DummyCreateHook(), ent.OpCreate),
		// Execute DummyBulkUpdateHook only for bulk Update operations
		hook.On(DummyBulkUpdateHook(), ent.OpUpdate),
		// Execute DummySingleUpdateHook only for single UpdateOne operations
		hook.On(DummySingleUpdateHook(), ent.OpUpdateOne),
		// Execute DummyBulkDeleteHook only for bulk Delete operations
		hook.On(DummyBulkDeleteHook(), ent.OpDelete),
		// Execute DummySingleDeleteHook only for single DeleteOne operations
		hook.On(DummySingleDeleteHook(), ent.OpDeleteOne),
	}
}
