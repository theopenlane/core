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

// ControlImplementation holds the schema definition for the ControlImplementation entity
type ControlImplementation struct {
	CustomSchema

	ent.Schema
}

const SchemaImplementation = "control_implementation"

func (ControlImplementation) Name() string {
	return SchemaImplementation
}

func (ControlImplementation) GetType() any {
	return ControlImplementation.Type
}

func (ControlImplementation) PluralName() string {
	return pluralize.NewClient().Plural(SchemaImplementation)
}

// Fields of the ControlImplementation
func (ControlImplementation) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			GoType(enums.DocumentStatus("")).
			Default(enums.DocumentDraft.String()).
			Optional().
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Comment("status of the %s, e.g. draft, published, archived, etc."),
		field.Time("implementation_date").
			Optional().
			Annotations(
				entgql.OrderField("implementation_date"),
			).
			Comment("date the control was implemented"),
		field.Bool("verified").
			Optional().
			Annotations(
				entgql.OrderField("verified"),
			).
			Comment("set to true if the control implementation has been verified"),
		field.Time("verification_date").
			Optional().
			Annotations(
				entgql.OrderField("verification_date"),
			).
			Comment("date the control implementation was verified"),
		field.Text("details").
			Optional().
			Comment("details of the control implementation"),
	}
}

// Mixin of the ControlImplementation
func (ControlImplementation) Mixin() []ent.Mixin {
	return getDefaultMixins()
}

// Edges of the ControlImplementation
func (c ControlImplementation) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Control{}),
	}
}

// Annotations of the ControlImplementation
func (ControlImplementation) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Policy of the ControlImplementation
func (ControlImplementation) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(), // TODO(sfunk): - add query rules
		),
	)
}
