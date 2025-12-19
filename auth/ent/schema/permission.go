package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
)

// Permission holds the schema definition for the Permission entity.
type Permission struct {
	ent.Schema
}

func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schema.BaseMixin{},
		schema.TenantMixin{},
	}
}

// Fields of the Permission.
// @generate-mutation: true
// @generate-resolver: true
// @generate-grpc: true
func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			NotEmpty().
			MaxLen(100).
			Comment("Permission name (e.g., users.create, products.read)"),
		field.String("display_name").
			NotEmpty().
			MaxLen(150).
			Comment("Human-readable permission name"),
		field.String("description").
			Optional().
			MaxLen(500).
			Comment("Permission description"),
		field.String("resource").
			NotEmpty().
			MaxLen(50).
			Comment("Resource this permission applies to (e.g., users, products, orders)"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether the permission is active"),
	}
}

// Edges of the Permission.
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		// Permission has many RolePermissions (junction table)
		edge.To("role_permissions", RolePermission.Type).
			Annotations(entgql.RelayConnection()).
			Comment("Role-Permission associations for this permission"),
	}
}

// Indexes of the Permission.
func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
		index.Fields("is_active"),
		index.Fields("created_at"),
	}
}
