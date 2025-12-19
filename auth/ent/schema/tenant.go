package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
)

// Tenant holds the schema definition for the Tenant entity.
type Tenant struct {
	ent.Schema
}

func (Tenant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schema.BaseMixin{},
	}
}

// Fields of the Tenant.
func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("Tenant name").
			Annotations(entgql.OrderField("NAME")),
		field.String("slug").
			Unique().
			NotEmpty().
			MaxLen(50).
			Comment("URL-friendly tenant identifier").
			Annotations(entgql.OrderField("SLUG")),
		field.String("domain").
			Optional().
			MaxLen(255).
			Comment("Custom domain for the tenant").
			Annotations(entgql.OrderField("DOMAIN")),
		field.Text("description").
			Optional().
			Comment("Tenant description"),
		field.Enum("status").
			Values("active", "inactive", "suspended", "pending").
			Default("pending").
			Comment("Tenant status").
			Annotations(entgql.OrderField("STATUS")),
		field.JSON("settings", map[string]interface{}{}).
			Optional().
			Comment("Tenant-specific configuration settings"),
		field.JSON("metadata", map[string]interface{}{}).
			Optional().
			Comment("Additional tenant metadata"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("Tenant expiration date (for trial/temporary tenants)").
			Annotations(entgql.OrderField("EXPIRES_AT")),
		field.Bool("is_active").
			Default(true).
			Comment("Whether the tenant is currently active").
			Annotations(entgql.OrderField("IS_ACTIVE")),
	}
}

// Edges of the Tenant.
func (Tenant) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Tenant.
func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").Unique(),
		index.Fields("domain").Unique(),
		index.Fields("status"),
		index.Fields("is_active"),
		index.Fields("created_at"),
	}
}
