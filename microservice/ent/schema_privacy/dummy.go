package privacy

import (
	"context"
	"github.com/saurabh/entgo-microservices/microservice/internal/ent"
	entprivacy "github.com/saurabh/entgo-microservices/microservice/internal/ent/privacy"

	"github.com/saurabh/entgo-microservices/pkg/authz"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func AllowIfBypassDummy() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		logger.WithFields(map[string]interface{}{
			"entity":     "Dummy",
			"rule":       "bypass",
			"query_type": q,
		}).Debug("Privacy rule evaluating")

		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "bypass",
			}).Debug("Error checking bypass")
			return entprivacy.Deny
		}
		if status == "Allow" {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "bypass",
				"result": "allow",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "bypass",
				"result": "deny",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}
		logger.WithFields(map[string]interface{}{
			"entity": "Dummy",
			"rule":   "bypass",
			"result": "skip",
		}).Debug("Privacy rule result")
		return entprivacy.Skip
	})
}

func HasRoleOrPermissionDummy() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		logger.WithFields(map[string]interface{}{
			"entity":     "Dummy",
			"rule":       "role_permission",
			"query_type": q,
		}).Debug("Privacy rule evaluating")

		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "role_permission",
				"result": "deny",
				"reason": "no_user_context",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}

		if authz.HasAnyRole(ctx, []string{"admin"}) {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "role_permission",
				"result": "allow",
				"check":  "role",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}

		if authz.HasPermission(ctx, "user", "can_read") {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "role_permission",
				"result": "allow",
				"check":  "permission",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}

		logger.WithFields(map[string]interface{}{
			"entity": "Dummy",
			"rule":   "role_permission",
			"result": "deny",
			"reason": "insufficient_privileges",
		}).Debug("Privacy rule result")
		return entprivacy.Deny
	})
}

func FilterByDummy() entprivacy.DummyQueryRuleFunc {
	return func(ctx context.Context, q *ent.DummyQuery) error {

		// No filtering annotations - skip filtering
		logger.WithFields(map[string]interface{}{
			"entity": "Dummy",
			"rule":   "filter",
			"result": "skip",
			"reason": "no_filter_annotations",
		}).Debug("Privacy rule result")
		return entprivacy.Skip

		logger.WithFields(map[string]interface{}{
			"entity": "Dummy",
			"rule":   "filter",
			"result": "allow",
		}).Debug("Privacy rule result")
		return entprivacy.Allow
	}
}

func AllowIfBypassDummyMutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logger.WithFields(map[string]interface{}{
			"entity":        "Dummy",
			"rule":          "bypass_mutation",
			"mutation_type": m,
		}).Debug("Privacy rule evaluating")

		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "bypass_mutation",
			}).Debug("Error checking bypass")
			return entprivacy.Deny
		}
		if status == "Allow" {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "bypass_mutation",
				"result": "allow",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "bypass_mutation",
				"result": "deny",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}
		logger.WithFields(map[string]interface{}{
			"entity": "Dummy",
			"rule":   "bypass_mutation",
			"result": "skip",
		}).Debug("Privacy rule result")
		return entprivacy.Skip
	})
}

func HasRoleOrPermissionDummyMutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logger.WithFields(map[string]interface{}{
			"entity":        "Dummy",
			"rule":          "role_permission_mutation",
			"mutation_type": m,
			"operation":     m.Op(),
		}).Debug("Privacy rule evaluating")

		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{
				"entity": "Dummy",
				"rule":   "role_permission_mutation",
				"result": "deny",
				"reason": "no_user_context",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}

		switch m.Op() {
		case ent.OpCreate:
			if authz.HasAnyRole(ctx, []string{"admin"}) {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"rule":      "role_permission_mutation",
					"operation": "create",
					"result":    "allow",
					"check":     "role",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "user", "can_create") {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"rule":      "role_permission_mutation",
					"operation": "create",
					"result":    "allow",
					"check":     "permission",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}

		case ent.OpUpdate, ent.OpUpdateOne:
			if authz.HasAnyRole(ctx, []string{"admin"}) {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"rule":      "role_permission_mutation",
					"operation": "update",
					"result":    "allow",
					"check":     "role",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "user", "can_update") {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"rule":      "role_permission_mutation",
					"operation": "update",
					"result":    "allow",
					"check":     "permission",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}

		case ent.OpDelete, ent.OpDeleteOne:
			if authz.HasAnyRole(ctx, []string{"admin"}) {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"rule":      "role_permission_mutation",
					"operation": "delete",
					"result":    "allow",
					"check":     "role",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "user", "can_delete") {
				logger.WithFields(map[string]interface{}{
					"entity":    "Dummy",
					"rule":      "role_permission_mutation",
					"operation": "delete",
					"result":    "allow",
					"check":     "permission",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
		}

		logger.WithFields(map[string]interface{}{
			"entity":    "Dummy",
			"rule":      "role_permission_mutation",
			"operation": m.Op(),
			"result":    "deny",
			"reason":    "insufficient_privileges",
		}).Debug("Privacy rule result")
		return entprivacy.Deny
	})
}

// DummyPolicy returns the complete privacy policy for Dummy
func DummyPolicy() ent.Policy {
	return entprivacy.Policy{
		Query: entprivacy.QueryPolicy{
			AllowIfBypassDummy(),
			HasRoleOrPermissionDummy(),
			FilterByDummy(),
		},
		Mutation: entprivacy.MutationPolicy{
			AllowIfBypassDummyMutation(),
			HasRoleOrPermissionDummyMutation(),
		},
	}
}
