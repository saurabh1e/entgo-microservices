#!/bin/bash

# Generate dummy schema with all annotations
generate_dummy_schema() {
    local service_dir=$1

    cat > "$service_dir/ent/schema/dummy.go" << 'EOF'
package schema

import (
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
func (Dummy) Policy() ent.Policy { return nil }

// Hooks is a placeholder for ent hooks that codegen may inject.
func (Dummy) Hooks() []ent.Hook { return nil }
EOF
}

# Copy base_mixin from auth or create minimal version
setup_base_mixin() {
    local service_dir=$1
    local auth_dir=$2

    if [ -f "$auth_dir/ent/schema/base_mixin.go" ]; then
        cp "$auth_dir/ent/schema/base_mixin.go" "$service_dir/ent/schema/"
        return 0
    fi

    # Create minimal base_mixin if auth doesn't have one
    cat > "$service_dir/ent/schema/base_mixin.go" << 'EOF'
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
// common fields with all package schemas.
type BaseMixin struct {
    mixin.Schema
}

func (BaseMixin) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("id").
            Positive().
            Immutable().
            Comment("Primary key").
            Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
                entgql.OrderField("ID"),
            ),
        field.Time("created_at").
            Immutable().
            Default(time.Now).
            Comment("Timestamp of entity creation").
            Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
                entgql.OrderField("CREATED_AT"),
            ),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now).
            Comment("Timestamp of last update").
            Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
                entgql.OrderField("UPDATED_AT"),
            ),
        field.Int64("tenant_id").
            Optional().
            Nillable().
            Comment("Tenant ID for multi-tenancy isolation").
            Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput)),
        field.Int64("created_by").
            Optional().
            Nillable().
            Comment("User ID who created this record").
            Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput)),
        field.Int64("owned_by").
            Optional().
            Nillable().
            Comment("User ID who owns this record").
            Annotations(entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput)),
    }
}

func (BaseMixin) Annotations() []schema.Annotation {
    return []schema.Annotation{
        entgql.RelayConnection(),
        entgql.QueryField(),
        entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
    }
}
EOF
}

