package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
)

// ControlImplementation holds the schema definition for the ControlImplementation entity
type ControlImplementation struct {
	SchemaFuncs

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
func (c ControlImplementation) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			// subcontrols can inherit permissions from the parent control
			newObjectOwnedMixin[generated.ControlImplementation](c,
				withParents(Control{}, Subcontrol{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins()
}

// Edges of the ControlImplementation
func (c ControlImplementation) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Control{}),
		defaultEdgeFromWithPagination(c, Subcontrol{}),
	}
}

// Hooks of the ControlImplementation
func (ControlImplementation) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookControlImplementation(),
	}
}

// Annotations of the ControlImplementation
func (ControlImplementation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ControlImplementation
func (ControlImplementation) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ControlImplementationMutation](rule.ControlsParent),    // if mutation contains control_id, check access
			rule.CanCreateObjectsUnderParent[*generated.ControlImplementationMutation](rule.SubcontrolsParent), // if mutation contains subcontrol_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlImplementationMutation](),
		),
	)
}
