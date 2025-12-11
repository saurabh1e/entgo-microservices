package main

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/saurabh/entgo-microservices/microservice/cmd/common"
)

const privacyTemplate = `package privacy

import (
	"context"
	"{{.ModuleName}}/internal/ent"{{if or .TenantIsolated .UserOwned .FilterByCreator}}
	"{{.ModuleName}}/internal/ent/{{.EntPackageName}}"{{end}}
	entprivacy "{{.ModuleName}}/internal/ent/privacy"

	"github.com/saurabh/entgo-microservices/pkg/authz"
	pkgcontext "github.com/saurabh/entgo-microservices/pkg/context"
	"github.com/saurabh/entgo-microservices/pkg/logger"
)

func AllowIfBypass{{.Entity}}() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "bypass",
			"query_type": q,
		}).Debug("Privacy rule evaluating")
		
		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "bypass",
			}).Debug("Error checking bypass")
			return entprivacy.Deny
		}
		if status == "Allow" {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "bypass",
				"result": "allow",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "bypass",
				"result": "deny",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "bypass",
			"result": "skip",
		}).Debug("Privacy rule result")
		return entprivacy.Skip
	})
}

func HasRoleOrPermission{{.Entity}}() entprivacy.QueryRule {
	return entprivacy.QueryRuleFunc(func(ctx context.Context, q ent.Query) error {
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "role_permission",
			"query_type": q,
		}).Debug("Privacy rule evaluating")

		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "role_permission",
				"result": "deny",
				"reason": "no_user_context",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}

		if authz.HasAnyRole(ctx, []string{ROLES_LIST_PLACEHOLDER}) {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "role_permission",
				"result": "allow",
				"check": "role",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}

		if authz.HasPermission(ctx, "{{.PermissionLevel}}", "can_read") {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "role_permission",
				"result": "allow",
				"check": "permission",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}

		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "role_permission",
			"result": "deny",
			"reason": "insufficient_privileges",
		}).Debug("Privacy rule result")
		return entprivacy.Deny
	})
}

func FilterBy{{.Entity}}() entprivacy.{{.Entity}}QueryRuleFunc {
	return func(ctx context.Context, q *ent.{{.Entity}}Query) error {
		{{if .TenantIsolated}}
		// Tenant isolation enabled via @tenant-isolated annotation
		tenantID, hasTenant := authz.GetTenantIDFromContext(ctx)
		if hasTenant {
			q.Where({{.EntPackageName}}.TenantIDEQ(tenantID))
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "filter",
				"filter_type": "tenant",
				"tenant_id": tenantID,
			}).Debug("Applied tenant filter")
		}
		{{end}}
		{{if .UserOwned}}
		// User ownership filtering enabled via @user-owned annotation
		userID, hasUser := authz.GetUserIDFromContext(ctx)
		if hasUser {
			q.Where({{.EntPackageName}}.OwnedByEQ(userID))
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "filter",
				"filter_type": "user_owned",
				"user_id": userID,
			}).Debug("Applied user ownership filter")
		}
		{{end}}
		{{if .FilterByCreator}}
		// Creator filtering enabled via @filter-by-creator annotation
		userID, hasUser := authz.GetUserIDFromContext(ctx)
		if hasUser {
			q.Where({{.EntPackageName}}.CreatedByEQ(userID))
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "filter",
				"filter_type": "creator",
				"user_id": userID,
			}).Debug("Applied creator filter")
		}
		{{end}}
		{{if not .TenantIsolated}}{{if not .UserOwned}}{{if not .FilterByCreator}}
		// No filtering annotations - skip filtering
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "filter",
			"result": "skip",
			"reason": "no_filter_annotations",
		}).Debug("Privacy rule result")
		return entprivacy.Skip
		{{end}}{{end}}{{end}}

		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "filter",
			"result": "allow",
		}).Debug("Privacy rule result")
		return entprivacy.Allow
	}
}

func AllowIfBypass{{.Entity}}Mutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "bypass_mutation",
			"mutation_type": m,
		}).Debug("Privacy rule evaluating")
		
		status, err := authz.CheckBypass(ctx)
		if err != nil {
			logger.WithError(err).WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "bypass_mutation",
			}).Debug("Error checking bypass")
			return entprivacy.Deny
		}
		if status == "Allow" {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "bypass_mutation",
				"result": "allow",
			}).Debug("Privacy rule result")
			return entprivacy.Allow
		}
		if status == "Deny" {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "bypass_mutation",
				"result": "deny",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "bypass_mutation",
			"result": "skip",
		}).Debug("Privacy rule result")
		return entprivacy.Skip
	})
}

func HasRoleOrPermission{{.Entity}}Mutation() entprivacy.MutationRule {
	return entprivacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "role_permission_mutation",
			"mutation_type": m,
			"operation": m.Op(),
		}).Debug("Privacy rule evaluating")

		_, ok := pkgcontext.GetUser(ctx)
		if !ok {
			logger.WithFields(map[string]interface{}{
				"entity": "{{.Entity}}",
				"rule": "role_permission_mutation",
				"result": "deny",
				"reason": "no_user_context",
			}).Debug("Privacy rule result")
			return entprivacy.Deny
		}

		switch m.Op() {
		case ent.OpCreate:
			if authz.HasAnyRole(ctx, []string{CREATE_ROLES_LIST_PLACEHOLDER}) {
				logger.WithFields(map[string]interface{}{
					"entity": "{{.Entity}}",
					"rule": "role_permission_mutation",
					"operation": "create",
					"result": "allow",
					"check": "role",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "{{.PermissionLevel}}", "can_create") {
				logger.WithFields(map[string]interface{}{
					"entity": "{{.Entity}}",
					"rule": "role_permission_mutation",
					"operation": "create",
					"result": "allow",
					"check": "permission",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}

		case ent.OpUpdate, ent.OpUpdateOne:
			if authz.HasAnyRole(ctx, []string{UPDATE_ROLES_LIST_PLACEHOLDER}) {
				logger.WithFields(map[string]interface{}{
					"entity": "{{.Entity}}",
					"rule": "role_permission_mutation",
					"operation": "update",
					"result": "allow",
					"check": "role",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "{{.PermissionLevel}}", "can_update") {
				logger.WithFields(map[string]interface{}{
					"entity": "{{.Entity}}",
					"rule": "role_permission_mutation",
					"operation": "update",
					"result": "allow",
					"check": "permission",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}

		case ent.OpDelete, ent.OpDeleteOne:
			if authz.HasAnyRole(ctx, []string{DELETE_ROLES_LIST_PLACEHOLDER}) {
				logger.WithFields(map[string]interface{}{
					"entity": "{{.Entity}}",
					"rule": "role_permission_mutation",
					"operation": "delete",
					"result": "allow",
					"check": "role",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
			if authz.HasPermission(ctx, "{{.PermissionLevel}}", "can_delete") {
				logger.WithFields(map[string]interface{}{
					"entity": "{{.Entity}}",
					"rule": "role_permission_mutation",
					"operation": "delete",
					"result": "allow",
					"check": "permission",
				}).Debug("Privacy rule result")
				return entprivacy.Allow
			}
		}

		logger.WithFields(map[string]interface{}{
			"entity": "{{.Entity}}",
			"rule": "role_permission_mutation",
			"operation": m.Op(),
			"result": "deny",
			"reason": "insufficient_privileges",
		}).Debug("Privacy rule result")
		return entprivacy.Deny
	})
}

// {{.Entity}}Policy returns the complete privacy policy for {{.Entity}}
func {{.Entity}}Policy() ent.Policy {
	return entprivacy.Policy{
		Query: entprivacy.QueryPolicy{
			AllowIfBypass{{.Entity}}(),
			HasRoleOrPermission{{.Entity}}(),
			FilterBy{{.Entity}}(),
		},
		Mutation: entprivacy.MutationPolicy{
			AllowIfBypass{{.Entity}}Mutation(),
			HasRoleOrPermission{{.Entity}}Mutation(),
		},
	}
}
`

type EntityData struct {
	Entity          string
	EntityLower     string
	EntPackageName  string
	ModuleName      string
	RolesList       string
	CreateRolesList string
	UpdateRolesList string
	DeleteRolesList string
	PermissionLevel string
	TenantIsolated  bool
	UserOwned       bool
	FilterByCreator bool
}

// Removed duplicate AnnotationData - now using common.AnnotationData

// formatRolesList converts a slice of roles to a quoted, comma-separated string
func formatRolesList(roles []string) string {
	if len(roles) == 0 {
		return `"admin", "manager", "user"` // default fallback
	}

	quotedRoles := make([]string, len(roles))
	for i, role := range roles {
		quotedRoles[i] = fmt.Sprintf(`"%s"`, role)
	}
	return strings.Join(quotedRoles, ", ")
}

// getCreateUpdateDeleteRoles returns role lists for different operations
func getCreateUpdateDeleteRoles(allRoles []string) (create, update, delete string) {
	if len(allRoles) == 0 {
		// Default behavior
		return `"admin", "manager"`, `"admin", "manager"`, `"admin"`
	}

	// For create and update, allow all roles except the most restrictive
	if len(allRoles) == 1 {
		create = formatRolesList(allRoles)
		update = formatRolesList(allRoles)
		delete = formatRolesList(allRoles)
	} else {
		// Create/Update: all roles except last (assuming last is most permissive)
		createUpdateRoles := allRoles[:len(allRoles)-1]
		create = formatRolesList(createUpdateRoles)
		update = formatRolesList(createUpdateRoles)

		// Delete: only first role (assuming first is most restrictive)
		delete = formatRolesList(allRoles[:1])
	}

	return create, update, delete
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Privacy generation failed: %v", err)
	}
	log.Println("Privacy files generation completed!")
}

func run() error {
	schemaDir := "ent/schema"
	privacyDir := "ent/schema_privacy"

	// Ensure privacy directory exists using common utility
	if err := common.EnsureDir(privacyDir); err != nil {
		return fmt.Errorf("creating privacy directory: %w", err)
	}

	// Get schema files using common utility
	schemaFiles, err := common.GetSchemaFiles(schemaDir)
	if err != nil {
		return fmt.Errorf("reading schema files: %w", err)
	}

	tmpl, err := template.New("privacy").Parse(privacyTemplate)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// Get the module name using common utility
	moduleName, err := common.GetModuleName()
	if err != nil {
		return fmt.Errorf("getting module name: %w", err)
	}

	for _, filePath := range schemaFiles {
		// Check if the schema file has the @generate-privacy: true annotation
		hasAnnotation, err := common.CheckAnnotation(filePath, "@generate-privacy: true")
		if err != nil {
			return fmt.Errorf("checking annotation for %s: %w", filePath, err)
		}

		if !hasAnnotation {
			log.Printf("Skipping %s - no @generate-privacy: true annotation found", filePath)
			continue
		}

		// Extract the actual entity name using common utility
		entity, err := common.ExtractEntityName(filePath)
		if err != nil {
			return fmt.Errorf("extracting entity name from %s: %w", filePath, err)
		}

		// Convert entity name to lowercase for permission checks
		entityLower := common.ToLowerSnakeCase(entity)

		// Parse role and permission annotations using common utility
		annotations, err := common.ParseAnnotations(filePath)
		if err != nil {
			return fmt.Errorf("parsing annotations for %s: %w", filePath, err)
		}

		// Set default permission level if not specified
		permissionLevel := annotations.PermissionLevel
		if permissionLevel == "" {
			permissionLevel = entityLower
		}

		// Parse roles from RoleLevel
		roles := common.ParseRolesList(annotations.RoleLevel)

		// Format roles list for template
		rolesList := formatRolesList(roles)
		createRolesList, updateRolesList, deleteRolesList := getCreateUpdateDeleteRoles(roles)

		data := EntityData{
			Entity:          entity,
			EntityLower:     entityLower,
			EntPackageName:  common.ToEntPackageName(entity),
			ModuleName:      moduleName,
			RolesList:       rolesList,
			CreateRolesList: createRolesList,
			UpdateRolesList: updateRolesList,
			DeleteRolesList: deleteRolesList,
			PermissionLevel: permissionLevel,
			TenantIsolated:  annotations.TenantIsolated,
			UserOwned:       annotations.UserOwned,
			FilterByCreator: annotations.FilterByEdge == "creator",
		}

		// Generate privacy file using template
		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			log.Fatalf("Failed to execute template for %s: %v", entity, err)
		}

		// Replace the role list placeholders with actual values - order matters!
		content := buf.String()
		// Replace longer placeholders first to avoid partial matches
		content = strings.ReplaceAll(content, "CREATE_ROLES_LIST_PLACEHOLDER", data.CreateRolesList)
		content = strings.ReplaceAll(content, "UPDATE_ROLES_LIST_PLACEHOLDER", data.UpdateRolesList)
		content = strings.ReplaceAll(content, "DELETE_ROLES_LIST_PLACEHOLDER", data.DeleteRolesList)
		content = strings.ReplaceAll(content, "ROLES_LIST_PLACEHOLDER", data.RolesList)

		// Debug: Show what we're replacing
		log.Printf("Debug - Roles: %s", data.RolesList)
		log.Printf("Debug - Create: %s", data.CreateRolesList)
		log.Printf("Debug - Update: %s", data.UpdateRolesList)
		log.Printf("Debug - Delete: %s", data.DeleteRolesList)

		// Debug: Print the content before formatting
		log.Printf("Generated content for %s:\n%s", entity, content)

		// Format the generated code using common utility
		formatted, err := common.FormatGoCode([]byte(content))
		if err != nil {
			return fmt.Errorf("formatting code for %s: %w", entity, err)
		}

		// Write to file using common utility
		outputFile := privacyDir + "/" + common.ToLowerSnakeCase(entity) + ".go"

		// SAFETY CHECK: Skip if file already exists to preserve custom code
		if common.FileExists(outputFile) {
			log.Printf("⚠️  SKIPPING %s - File already exists. Delete to regenerate.", outputFile)
			continue
		}

		if err := common.WriteGoFile(outputFile, formatted); err != nil {
			return fmt.Errorf("writing privacy file for %s: %w", entity, err)
		}

		log.Printf("✅ Generated privacy file: %s", outputFile)
	}

	return nil
}
