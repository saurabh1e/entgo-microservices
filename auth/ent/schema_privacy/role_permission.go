package privacy

import (
	"context"
	"github.com/saurabh/entgo-microservices/auth/internal/ent"
	entprivacy "github.com/saurabh/entgo-microservices/auth/internal/ent/privacy"
	"github.com/saurabh/entgo-microservices/auth/internal/ent/rolepermission"

	"github.com/saurabh/entgo-microservices/pkg/authz"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func AllowIfBypassRolePermission() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "bypass"}).Warn("Bypass check error")
			return entprivacy.Deny
		}
		if status == "Allow" {
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "bypass"}).Warn("Bypass explicitly denied")
			return entprivacy.Deny
		}
		return entprivacy.Skip
	})
}

func HasRoleOrPermissionRolePermission() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "role_permission"}).Warn("No user in context - denying access")
			return entprivacy.Deny
		}

		if authz.HasAnyRole(ctx, []string{"admin"}) {
			return entprivacy.Skip
		}

		if authz.HasPermission(ctx, "user", "can_read") {
			return entprivacy.Skip
		}

		logger.WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "role_permission"}).Warn("Insufficient privileges - denying access")
		return entprivacy.Deny
	})
}

func FilterByRolePermission() entprivacy.RolePermissionQueryRuleFunc {
	return func(ctx context.Context, q *ent.RolePermissionQuery) error {
		applied := false

		// Tenant isolation: apply tenant filter if tenant info is available
		tenantID, err := pkgcontext.GetUserTenantID(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{"entity": "RolePermission", "filter": "tenant"}).Error("Failed to get tenant ID from context - denying access")
			return entprivacy.Deny
		}

		q.Where(rolepermission.TenantIDEQ(tenantID))
		applied = true
		logger.WithFields(map[string]interface{}{"entity": "RolePermission", "filter": "tenant", "tenant_id": tenantID}).Info("Applied tenant filter")

		if !applied {
			// No filters applied - skip this rule
			return entprivacy.Skip
		}

		return entprivacy.Allow
	}
}

func AllowIfBypassRolePermissionMutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "bypass_mutation"}).Warn("Bypass mutation check error")
			return entprivacy.Deny
		}
		if status == "Allow" {
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "bypass_mutation"}).Warn("Bypass mutation explicitly denied")
			return entprivacy.Deny
		}
		return entprivacy.Skip
	})
}

func HasRoleOrPermissionRolePermissionMutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "role_permission_mutation"}).Warn("No user in context - denying mutation")
			return entprivacy.Deny
		}

		// For update/delete operations, validate tenant ID matches context
		if m.Op() == ent.OpUpdate || m.Op() == ent.OpUpdateOne || m.Op() == ent.OpDelete || m.Op() == ent.OpDeleteOne {
			contextTenantID, err := pkgcontext.GetUserTenantID(ctx)
			if err != nil {
				logger.WithError(err).WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "tenant_validation", "operation": m.Op()}).Error("Failed to get tenant ID from context")
				return entprivacy.Deny
			}

			if rolepermissionMutation, ok := m.(*ent.RolePermissionMutation); ok {
				if tenantID, exists := rolepermissionMutation.TenantID(); exists && tenantID != contextTenantID {
					logger.WithFields(map[string]interface{}{
						"entity":            "RolePermission",
						"rule":              "tenant_validation",
						"operation":         m.Op(),
						"context_tenant_id": contextTenantID,
						"record_tenant_id":  tenantID,
					}).Warn("Tenant ID mismatch - denying mutation")
					return entprivacy.Deny
				}
			}
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

		logger.WithFields(map[string]interface{}{"entity": "RolePermission", "rule": "role_permission_mutation", "operation": m.Op()}).Warn("Insufficient privileges for mutation - denying")
		return entprivacy.Deny
	})
}

// RolePermissionPolicy returns the complete privacy policy for RolePermission
func RolePermissionPolicy() ent.Policy {
	return entprivacy.Policy{
		Query: entprivacy.QueryPolicy{
			AllowIfBypassRolePermission(),
			HasRoleOrPermissionRolePermission(),
			FilterByRolePermission(),
		},
		Mutation: entprivacy.MutationPolicy{
			AllowIfBypassRolePermissionMutation(),
			HasRoleOrPermissionRolePermissionMutation(),
		},
	}
}
