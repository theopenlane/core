package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx/history"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx/accessmap"
)

// Onboarding holds the schema definition for the Onboarding entity
type Onboarding struct {
	SchemaFuncs

	ent.Schema
}

// SchemaOnboarding is the name of the Onboarding schema.
const SchemaOnboarding = "onboarding"

// Name returns the name of the Onboarding schema.
func (Onboarding) Name() string {
	return SchemaOnboarding
}

// GetType returns the type of the Onboarding schema.
func (Onboarding) GetType() any {
	return Onboarding.Type
}

// PluralName returns the plural name of the Onboarding schema.
func (Onboarding) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOnboarding)
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

func (Onboarding) Features() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
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
func (o Onboarding) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			edgeSchema: Organization{},
			field:      "organization_id",
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Annotations of the Onboarding
func (o Onboarding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Mutations(entgql.MutationCreate()),
		// don't store the history of the onboarding
		history.Annotations{
			Exclude: true,
		},
	}
}

// Interceptors of the Onboarding
func (o Onboarding) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorFeatures(o.Features()...),
	}
}

// Hooks of the Onboarding
func (Onboarding) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOnboarding(),
	}
}

// Policy of the Onboarding
func (o Onboarding) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithQueryRules(
			// this data should not be queried, so we deny all queries except
			// those explicitly allowed from internal services
			rule.AllowIfContextAllowRule(),
			privacy.AlwaysDenyRule(), // deny all queries by default
		),
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
			rule.DenyIfMissingAllFeatures("onboarding", o.Features()...),
			privacy.AlwaysAllowRule(), // Allow all other users (e.g. a user with a JWT should be able to create a new org)
		),
	)
}
