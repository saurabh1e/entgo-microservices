package privacy

import (
	"context"
	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	entprivacy "github.com/saurabh/entgo-microservices/auth/internal/ent/privacy"

	"github.com/saurabh/entgo-microservices/pkg/authz"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func AllowIfBypassRole() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{"entity": "Role", "rule": "bypass"}).Warn("Bypass check error")
			return entprivacy.Deny
		}
		if status == "Allow" {
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{"entity": "Role", "rule": "bypass"}).Warn("Bypass explicitly denied")
			return entprivacy.Deny
		}
		return entprivacy.Skip
	})
}

func HasRoleOrPermissionRole() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{"entity": "Role", "rule": "role_permission"}).Warn("No user in context - denying access")
			return entprivacy.Deny
		}

		if authz.HasAnyRole(ctx, []string{"admin"}) {
			return entprivacy.Skip
		}

		if authz.HasPermission(ctx, "user", "can_read") {
			return entprivacy.Skip
		}

		logger.WithFields(map[string]interface{}{"entity": "Role", "rule": "role_permission"}).Warn("Insufficient privileges - denying access")
		return entprivacy.Deny
	})
}

func FilterByRole() entprivacy.RoleQueryRuleFunc {
	return func(ctx context.Context, q *ent.RoleQuery) error {
		applied := false

		if !applied {
			// No filters applied - skip this rule
			return entprivacy.Skip
		}

		return entprivacy.Allow
	}
}

func AllowIfBypassRoleMutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{"entity": "Role", "rule": "bypass_mutation"}).Warn("Bypass mutation check error")
			return entprivacy.Deny
		}
		if status == "Allow" {
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{"entity": "Role", "rule": "bypass_mutation"}).Warn("Bypass mutation explicitly denied")
			return entprivacy.Deny
		}
		return entprivacy.Skip
	})
}

func HasRoleOrPermissionRoleMutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{"entity": "Role", "rule": "role_permission_mutation"}).Warn("No user in context - denying mutation")
			return entprivacy.Deny
		}

		switch m.Op() {
		case ent.OpCreate:
			if authz.HasAnyRole(ctx, []string{"admin"}) {
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "user", "can_create") {
				return entprivacy.Allow
			}

		case ent.OpUpdate, ent.OpUpdateOne:
			if authz.HasAnyRole(ctx, []string{"admin"}) {
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "user", "can_update") {
				return entprivacy.Allow
			}

		case ent.OpDelete, ent.OpDeleteOne:
			if authz.HasAnyRole(ctx, []string{"admin"}) {
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "user", "can_delete") {
				return entprivacy.Allow
			}
		}

		logger.WithFields(map[string]interface{}{"entity": "Role", "rule": "role_permission_mutation", "operation": m.Op()}).Warn("Insufficient privileges for mutation - denying")
		return entprivacy.Deny
	})
}

// RolePolicy returns the complete privacy policy for Role
func RolePolicy() ent.Policy {
	return entprivacy.Policy{
		Query: entprivacy.QueryPolicy{
			AllowIfBypassRole(),
			HasRoleOrPermissionRole(),
			FilterByRole(),
		},
		Mutation: entprivacy.MutationPolicy{
			AllowIfBypassRoleMutation(),
			HasRoleOrPermissionRoleMutation(),
		},
	}
}
