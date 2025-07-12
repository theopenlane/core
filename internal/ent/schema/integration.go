package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx"
)

// Integration maps configured integrations (github, slack, etc.) to organizations
type Integration struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIntegration is the name of the Integration schema.
const SchemaIntegration = "integration"

// Name returns the name of the Integration schema.
func (Integration) Name() string {
	return SchemaIntegration
}

// GetType returns the type of the Integration schema.
func (Integration) GetType() any {
	return Integration.Type
}

// PluralName returns the plural name of the Integration schema.
func (Integration) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegration)
}

// Fields of the Integration
func (Integration) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the integration - must be unique within the organization").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("a description of the integration").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("kind").
			Optional().
			Annotations(
				entgql.OrderField("kind"),
			),
	}
}

// Edges of the Integration
func (i Integration) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: i,
			edgeSchema: Hush{},
			comment:    "the secrets associated with the integration",
		}),
		defaultEdgeToWithPagination(i, Event{}),
	}
}

// Mixin of the Integration
func (i Integration) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(i),
		},
	}.getMixins()
}

// Policy of the Integration
func (Integration) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the Integration
func (Integration) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("base"),
	}
}
