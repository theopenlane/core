package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	emixin "github.com/theopenlane/entx/mixin"
)

// MappedControl holds the schema definition for the MappedControl entity
type MappedControl struct {
	ent.Schema
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
func (MappedControl) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controls", Control.Type).
			Annotations(entgql.RelayConnection()).
			Comment("mapped controls that have a relation to each other"),
		edge.To("subcontrols", Subcontrol.Type).
			Annotations(entgql.RelayConnection()).
			Comment("mapped subcontrols that have a relation to each other"),
	}
}

// Mixin of the MappedControl
func (MappedControl) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
	}
}

// Annotations of the MappedControl
func (MappedControl) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
	}
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
