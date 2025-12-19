package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// BaseMixin implements the ent.Mixin for sharing
// privacy fields with all package schemas.
type BaseMixin struct {
	mixin.Schema
}

// Fields of the BaseMixin.
func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Immutable().
			Comment("Primary key").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
				entgql.OrderField("ID"),
				entgql.Type("ID")),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Creation timestamp").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput | entgql.SkipWhereInput)),
		field.Int("created_by").
			Optional().
			Nillable().
			Comment("User ID who created this record").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
		field.Int("owned_by").
			Optional().
			Nillable().
			Comment("User ID who owns this record").
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput)),
	}
}

// Indexes of the BaseMixin.
func (BaseMixin) Indexes() []ent.Index {
	return []ent.Index{}
}

// Policy of the BaseMixin (optional, can be overridden by schemas)
func (BaseMixin) Policy() ent.Policy {
	return nil
}

func (BaseMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entgql.Directives(),
	}
}
