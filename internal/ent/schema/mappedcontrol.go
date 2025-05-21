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

// MappedControl holds the schema definition for the MappedControl entity
type MappedControl struct {
	SchemaFuncs

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
		field.Enum("mapping_type").
			GoType(enums.MappingType("")).
			Comment("the type of mapping between the two controls, e.g. subset, intersect, equal, superset").
			Annotations(
				entgql.OrderField("MAPPING_TYPE"),
			).
			Default(enums.MappingTypeEqual.String()).
			Optional(),
		field.String("relation").
			Comment("description of how the two controls are related").
			Optional(),
		field.String("confidence").
			Comment("percentage of confidence in the mapping").
			Optional(),
		field.Enum("source").
			GoType(enums.MappingSource("")).
			Optional().
			Annotations(
				entgql.OrderField("SOURCE"),
			).
			Default(enums.MappingSourceManual.String()).
			Comment("source of the mapping, e.g. manual, suggested, etc."),
	}
}

// Edges of the MappedControl
func (m MappedControl) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Control.Type,
			name:       "from_control",
			comment:    "controls that map to another control",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Control.Type,
			name:       "to_control",
			comment:    "controls that are being mapped from another control",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Subcontrol.Type,
			name:       "from_subcontrol",
			comment:    "subcontrols map to another control",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Subcontrol.Type,
			name:       "to_subcontrol",
			comment:    "subcontrols are being mapped from another control",
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
