package schema

import (
	"fmt"
	"regexp"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// CodeMixin implements the ent.Mixin for sharing
// auto-generated code field: tenant:<id>:code:lower(<name>.replace(" ", "_"))
// Code generation is handled by the generate-hooks command.
type CodeMixin struct {
	mixin.Schema
}

// Fields of the CodeMixin.
func (CodeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").
			Immutable().
			Comment("Auto-generated unique code identifier"),
	}
}

// GenerateCode creates the code from tenant_id and name
// Format: tenant:<id>:code:lower(<name>.replace(" ", "_"))
// This function is used by the generated hooks.
func GenerateCode(tenantID int, name string) string {
	// Convert to lowercase
	normalized := strings.ToLower(name)

	// Replace spaces with underscores
	normalized = strings.ReplaceAll(normalized, " ", "_")

	// Remove any characters that aren't alphanumeric, underscore, or hyphen
	reg := regexp.MustCompile(`[^a-z0-9_-]`)
	normalized = reg.ReplaceAllString(normalized, "")

	// Remove multiple consecutive underscores
	reg = regexp.MustCompile(`_+`)
	normalized = reg.ReplaceAllString(normalized, "_")

	// Trim leading/trailing underscores
	normalized = strings.Trim(normalized, "_")

	return fmt.Sprintf("tenant:%d:code:%s", tenantID, normalized)
}
