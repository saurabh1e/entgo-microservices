package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	hook "github.com/saurabh/entgo-microservices/auth/ent/schema_hooks"
	privacy "github.com/saurabh/entgo-microservices/auth/ent/schema_privacy"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
)

// RolePermission holds the schema definition for the RolePermission junction entity.
type RolePermission struct {
	ent.Schema
}

func (RolePermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schema.BaseMixin{},
		schema.TenantMixin{},
	}
}

// Fields of the RolePermission.
// @generate-resolver: true
// @generate-mutation: true
// @generate-hooks: true
// @generate-privacy: true
// @generate-grpc: true
// @role-level: admin
// @permission-level: user
// @tenant-isolated: true
func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("can_read").Default(false).Comment("Can read the resource"),
		field.Bool("can_create").Default(false).Comment("Can create the resource"),
		field.Bool("can_update").Default(false).Comment("Can update the resource"),
		field.Bool("can_delete").Default(false).Comment("Can delete the resource"),
	}
}

// Edges of the RolePermission.
func (RolePermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Role.Type).
			Unique().
			Required(),
		edge.To("permission", Permission.Type).
			Unique().
			Required(),
	}
}

// Indexes of the RolePermission.
func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		// Unique constraint handled by edge configuration
	}
}

func (RolePermission) Policy() ent.Policy {
	return privacy.RolePermissionPolicy()
}

func (RolePermission) Hooks() []ent.Hook {
	return hook.RolePermissionHooks()
}
