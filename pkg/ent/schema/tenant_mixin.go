package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// TenantMixin implements the ent.Mixin for sharing
// mandatory tenant_id field with all schemas except Tenant itself.
type TenantMixin struct {
	mixin.Schema
}

// Fields of the TenantMixin.
func (TenantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("tenant_id").
			Positive().
			Comment("Tenant ID for multi-tenancy isolation").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
	}
}
