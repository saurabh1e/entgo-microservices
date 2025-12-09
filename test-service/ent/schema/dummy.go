package schema

import (
	hook "github.com/saurabh/entgo-microservices/test-service/ent/schema_hooks"
	privacy "github.com/saurabh/entgo-microservices/test-service/ent/schema_privacy"
    "entgo.io/ent"
    "entgo.io/ent/schema/field"

    "entgo.io/contrib/entgql"
)

// Dummy holds the schema definition for the Dummy entity.
// @generate-resolver: true
// @generate-mutation: true
// @generate-hooks: true
// @generate-privacy: true
// @role-level: admin
// @permission-level: user
// @generate-grpc: true
type Dummy struct {
    ent.Schema
}

// Mixin returns the list of mixins for the Dummy entity.
func (Dummy) Mixin() []ent.Mixin {
    return []ent.Mixin{
        BaseMixin{},
    }
}

// Fields of the Dummy.
func (Dummy) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").Optional().Nillable().Comment("Dummy name").
            Annotations(
                entgql.OrderField("NAME"),
                entgql.Type("String"),
            ),
    }
}

// Edges of the Dummy.
func (Dummy) Edges() []ent.Edge { return nil }

// Indexes returns indexes for the Dummy schema.
func (Dummy) Indexes() []ent.Index { return nil }

// Policy is a placeholder for privacy policy that codegen may inject.

func (Dummy) Policy() ent.Policy {
	return privacy.DummyPolicy()
}

func (Dummy) Hooks() []ent.Hook {
	return hook.DummyHooks()
}
