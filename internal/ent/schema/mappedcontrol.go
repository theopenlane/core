package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// MappedControl holds the schema definition for the MappedControl entity
type MappedControl struct {
	CustomSchema

	ent.Schema
}

const SchemaMappedControl = "mapped_control"

func (MappedControl) Name() string {
	return SchemaMappedControl
}

func (MappedControl) GetType() any {
	return MappedControl.Type
}

func (MappedControl) PluralName() string {
	return pluralize.NewClient().Plural(SchemaMappedControl)
}

// Fields of the MappedControl
func (MappedControl) Fields() []ent.Field {
	return []ent.Field{
		field.String("mapping_type").
			Comment("the type of mapping between the two controls, e.g. subset, intersect, equal, superset").
			Annotations(
				entgql.OrderField("mapping_type"),
			).
			Optional(),
		field.String("relation").
			Comment("description of how the two controls are related").
			Optional(),
	}
}

// Edges of the MappedControl
func (m MappedControl) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Control{},
			comment:    "mapped controls that have a relation to each other",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Subcontrol{},
			comment:    "mapped subcontrols that have a relation to each other",
		}),
	}
}

// Mixin of the MappedControl
func (MappedControl) Mixin() []ent.Mixin {
	return getDefaultMixins()

}

// Annotations of the MappedControl
func (MappedControl) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Hooks of the MappedControl
func (MappedControl) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the MappedControl
func (MappedControl) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the MappedControl
func (MappedControl) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
	)
}
