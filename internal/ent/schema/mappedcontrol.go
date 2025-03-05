package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

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
		field.String("control_id").
			Comment("the id of the control being mapped").
			Unique().
			Immutable().
			NotEmpty(),
		field.String("mapped_control_id").
			Unique().
			Immutable().
			NotEmpty().
			Comment("the id of the control that is mapped to"),
		field.String("mapping_type").
			Comment("the type of mapping between the two controls, e.g. subset, intersect, equal, superset").
			Optional(),
		field.String("relation").
			Comment("description of how the two controls are related").
			Optional(),
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

// Edges of the MappedControl
func (MappedControl) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("control", Control.Type).
			Unique().
			Required().
			Field("control_id").
			Immutable(),
		edge.To("mapped_control", Control.Type).
			Comment("mapped control to the original control, meaning there is overlap between the controls").
			Unique().
			Required().
			Field("mapped_control_id").
			Immutable(),
	}
}

// Indexes of the MappedControl
func (MappedControl) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("control_id", "mapped_control_id").
			Unique(),
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
