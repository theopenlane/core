package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// Usage holds per organization resource usage
type Usage struct {
	SchemaFuncs
	ent.Schema
}

const SchemaUsage = "usage"

func (Usage) Name() string       { return SchemaUsage }
func (Usage) GetType() any       { return Usage.Type }
func (Usage) PluralName() string { return pluralize.NewClient().Plural(SchemaUsage) }

// Fields defines the fields for the Usage schema
func (Usage) Fields() []ent.Field {
	return []ent.Field{
		field.String("organization_id").
			Comment("owner organization"),
		field.Enum("resource_type").
			Comment("the type of resource this usage is for").
			GoType(enums.UsageType("")),
		field.Int64("used").
			Comment("the amount of resource used").
			NonNegative().
			Default(0),
		field.Int64("limit").
			Comment("the limit for the resource type").
			NonNegative().
			Default(0),
	}
}

// Mixin returns the mixins for the Usage schema
func (Usage) Mixin() []ent.Mixin {
	return mixinConfig{excludeAnnotations: true}.getMixins()
}

// Annotations returns the annotations for the Usage schema
func (Usage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),

		entgql.QueryField(),

		entgql.Skip(entgql.SkipMutationCreateInput |
			entgql.SkipMutationUpdateInput),
	}
}

// Edges defines the edges for the Usage schema
func (Usage) Edges() []ent.Edge { return nil }

// Hooks defines the hooks for the Usage schema
func (Usage) Hooks() []ent.Hook { return nil }

// Policy returns the privacy policy for the Usage schema
func (Usage) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(),
		),
	)
}
