package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	hook "github.com/saurabh/entgo-microservices/auth/ent/schema_hooks"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
)

// Brand holds the schema definition for the Brand entity.
type Brand struct {
	ent.Schema
}

func (Brand) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schema.BaseMixin{},
		schema.TenantMixin{},
		schema.CodeMixin{}, // Auto-generates code from tenant_id + name
	}
}

// Fields of the Brand.
// @generate-resolver: true
// @generate-mutation: true
// @generate-grpc: true
// @generate-hooks: true
// @tenant-isolated: true
func (Brand) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Annotations(entgql.OrderField("NAME")).
			Comment("Brand name"),
	}
}

// Edges of the Brand.
func (Brand) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Brand.
func (Brand) Indexes() []ent.Index {
	return []ent.Index{
		// Unique brand code per tenant
		index.Fields("tenant_id", "code").Unique(),
		index.Fields("name"),
	}
}

// Hooks of the Brand.

func (Brand) Hooks() []ent.Hook {
	return hook.BrandHooks()
}
