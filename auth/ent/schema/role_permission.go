package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
// @generate-mutation: true
// @generate-resolver: true
// @generate-grpc: true
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
		// RolePermission belongs to one Role
		edge.From("role", Role.Type).
			Ref("role_permissions").
			Unique().
			Required().
			Comment("Role that this permission association belongs to"),
		// RolePermission belongs to one Permission
		edge.From("permission", Permission.Type).
			Ref("role_permissions").
			Unique().
			Required().
			Comment("Permission that this role association belongs to"),
	}
}

// Indexes of the RolePermission.
func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		// Ensure unique role-permission combination
	}
}
