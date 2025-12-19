package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	privacy "github.com/saurabh/entgo-microservices/auth/ent/schema_privacy"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schema.BaseMixin{},
		schema.TenantMixin{},
	}
}

// Fields of the Role.
// @generate-resolver: true
// @generate-mutation: true
// @generate-privacy: true
// @generate-grpc: true
// @role-level: admin
// @permission-level: user
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			NotEmpty().
			MaxLen(50).
			Annotations(entgql.OrderField("NAME")).
			Comment("Role name (e.g., admin, user, moderator)"),
		field.String("display_name").
			NotEmpty().
			MaxLen(100).
			Comment("Human-readable role name"),
		field.String("description").
			Optional().
			MaxLen(500).
			Comment("Role description"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether the role is active"),
		field.Int("priority").
			Default(0).
			Comment("Role priority for hierarchy (higher number = higher priority)"),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		// Role has many Users
		edge.To("users", User.Type).
			Comment("Users with this role").
			Annotations(entgql.RelayConnection()),

		// Role has many RolePermissions (junction table)
		edge.To("role_permissions", RolePermission.Type).
			Comment("Role-Permission associations for this role").
			Annotations(entgql.RelayConnection()),
	}
}

// Indexes of the Role.
func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
		index.Fields("is_active"),
		index.Fields("priority"),
		index.Fields("created_at"),
	}
}

func (Role) Policy() ent.Policy {
	return privacy.RolePolicy()
}
