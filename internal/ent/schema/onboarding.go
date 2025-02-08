package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"
)

// Onboarding holds the schema definition for the Onboarding entity
type Onboarding struct {
	ent.Schema
}

// Fields of the Onboarding
func (Onboarding) Fields() []ent.Field {
	return []ent.Field{
		field.String("organization_id").
			Unique().
			Optional().
			NotEmpty().
			Immutable(),
		field.String("company_name").
			Comment("name of the company"),
		field.Strings("domains").
			Comment("domains associated with the company").
			Optional(),
		field.JSON("company_details", map[string]any{}).
			Comment("details given about the company during the onboarding process, including things such as company size, sector, etc").
			Optional(),
		field.JSON("user_details", map[string]any{}).
			Comment("details given about the user during the onboarding process, including things such as name, job title, department, etc").
			Optional(),
		field.JSON("compliance", map[string]any{}).
			Comment("details given about the compliance requirements during the onboarding process, such as coming with existing policies, controls, risk assessments, etc").
			Optional(),
	}
}

// Mixin of the Onboarding
func (Onboarding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the Onboarding
func (Onboarding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Unique().
			Immutable().
			Field("organization_id"),
	}
}

// Indexes of the Onboarding
func (Onboarding) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Onboarding
func (Onboarding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Mutations(entgql.MutationCreate()),
		// don't store the history of the onboarding
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the Onboarding
func (Onboarding) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOnboarding(),
	}
}

// Interceptors of the Onboarding
func (Onboarding) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the Onboarding
func (Onboarding) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithQueryRules(
			// this data should not be queried, so we deny all queries except
			// those explicitly allowed from internal services
			rule.AllowIfContextAllowRule(),
		),
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
			privacy.AlwaysAllowRule(), // Allow all other users (e.g. a user with a JWT should be able to create a new org)
		),
	)
}
