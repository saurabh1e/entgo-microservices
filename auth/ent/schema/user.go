package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	hook "github.com/saurabh/entgo-microservices/auth/ent/schema_hooks"
	privacy "github.com/saurabh/entgo-microservices/auth/ent/schema_privacy"
	"github.com/saurabh/entgo-microservices/pkg/ent/schema"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schema.BaseMixin{},
		schema.TenantMixin{},
	}
}

// Fields of the User.
// @generate-resolver: true
// @generate-mutation: true
// @generate-hooks: true
// @generate-privacy: true
// @generate-grpc: true
// @role-level: admin
// @permission-level: user
// @tenant-isolated: true
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("username").
			Unique().
			NotEmpty().
			MinLen(3).
			MaxLen(50),
		field.String("password_hash").
			NotEmpty().
			Sensitive().
			Comment("Hashed password"),
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Annotations(entgql.OrderField("NAME")),
		field.String("phone").
			Optional().
			MaxLen(20),
		field.Text("address").
			Optional(),

		// User type field
		field.String("user_type").
			MaxLen(50).
			Default("staff").
			Comment("Type of user - admin, staff, vendor, or customer"),

		// Business-specific fields (used based on role)
		field.String("user_code").
			Optional().
			MaxLen(20).
			Unique().
			Comment("Vendor code or Customer code"),
		field.String("company_name").
			Optional().
			MaxLen(200).
			Comment("For vendors and corporate customers"),
		field.String("customer_type").
			Optional().
			MaxLen(50).
			Comment("For customers only - individual or corporate"),
		field.Int("payment_terms").
			Optional().
			Comment("For vendors only - payment days"),

		field.Bool("is_active").
			Default(true),
		field.Bool("email_verified").
			Default(false),
		field.Time("email_verified_at").
			Optional().
			Nillable(),
		field.Time("last_login").
			Optional().
			Nillable(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Role.Type).
			Unique(),
	}
}

func (User) Policy() ent.Policy {
	return privacy.UserPolicy()
}

func (User) Hooks() []ent.Hook {
	return hook.UserHooks()
}
