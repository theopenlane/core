package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Control defines the control schema.
type Control struct {
	SchemaFuncs

	ent.Schema
}

const SchemaControl = "control"

func (Control) Name() string {
	return SchemaControl
}

func (Control) GetType() any {
	return Control.Type
}

func (c Control) GetStringType() any {
	return Control.Type
}

func (Control) PluralName() string {
	return pluralize.NewClient().Plural(SchemaControl)
}

// Fields returns control fields.
func (Control) Fields() []ent.Field {
	// add any fields that are specific to the parent control here
	additionalFields := []ent.Field{
		field.String("ref_code").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("ref_code"),
			).
			NotEmpty().
			Comment("the unique reference code for the control"),
		field.String("standard_id").
			Comment("the id of the standard that the control belongs to, if applicable").
			Optional(),
	}

	return append(controlFields, additionalFields...)
}

// Edges of the Control
func (c Control) Edges() []ent.Edge {
	return []ent.Edge{
		// parents of the control (standard, program)
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Standard{},
			field:      "standard_id",
		}),
		defaultEdgeFromWithPagination(c, Program{}),

		defaultEdgeFromWithPagination(c, Evidence{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    c,
			edgeSchema:    ControlImplementation{},
			cascadeDelete: "Controls",
			comment:       "the implementation(s) of the control",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: MappedControl{},
			comment:    "mapped subcontrols that have a relation to another control or subcontrol",
		}),

		// controls have control objectives and subcontrols
		defaultEdgeToWithPagination(c, ControlObjective{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    c,
			edgeSchema:    Subcontrol{},
			cascadeDelete: "Control",
		}),

		// controls can have associated task, narratives, risks, and action plans
		defaultEdgeToWithPagination(c, Task{}),
		defaultEdgeToWithPagination(c, Narrative{}),
		defaultEdgeToWithPagination(c, Risk{}),
		defaultEdgeToWithPagination(c, ActionPlan{}),

		// policies and procedures are used to implement the control
		defaultEdgeToWithPagination(c, Procedure{}),
		defaultEdgeToWithPagination(c, InternalPolicy{}),
		// owner is the user who is responsible for the control
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "control_owner",
			t:          Group.Type,
			comment:    "the group of users who are responsible for the control, will be assigned tasks, approval, etc.",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "delegate",
			t:          Group.Type,
			comment:    "temporary delegate for the control, used for temporary control ownership",
		}),
	}
}

func (Control) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the standard
		index.Fields("standard_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
	}
}

// Mixin of the Control
func (c Control) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "CTL",
		additionalMixins: []ent.Mixin{
			// controls must be associated with an organization but do not inherit permissions from the organization
			// controls can inherit permissions from the associated programs
			newObjectOwnedMixin[generated.Control](c,
				withParents(Program{}, Standard{}),
				withOrganizationOwner(true),
			),
			// add groups permissions with viewer, editor, and blocked groups
			newGroupPermissionsMixin(),
		},
	}.getMixins()
}

// Annotations of the Control
func (Control) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the Control
func (Control) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ControlMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlMutation](),
		),
	)
}
