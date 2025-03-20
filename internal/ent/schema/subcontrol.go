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
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Subcontrol defines the file schema.
type Subcontrol struct {
	CustomSchema

	ent.Schema
}

const SchemaSubcontrol = "subcontrol"

func (Subcontrol) Name() string {
	return SchemaSubcontrol
}

func (Subcontrol) GetType() any {
	return Subcontrol.Type
}

func (Subcontrol) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSubcontrol)
}

// Fields returns file fields.
func (Subcontrol) Fields() []ent.Field {
	// add any fields that are specific to the subcontrol here
	additionalFields := []ent.Field{
		field.String("ref_code").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("ref_code"),
			).
			NotEmpty().
			Comment("the unique reference code for the control"),
		field.String("control_id").
			Unique().
			Comment("the id of the parent control").
			NotEmpty(),
	}

	return append(controlFields, additionalFields...)
}

// Edges of the Subcontrol
func (s Subcontrol) Edges() []ent.Edge {
	return []ent.Edge{
		// subcontrols are required to have a parent control
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Control{},
			field:      "control_id",
			required:   true,
		}),
		// controls can be mapped to other controls as a reference
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: MappedControl{},
			comment:    "mapped subcontrols that have a relation to another control or subcontrol",
		}),
		// evidence can be associated with the control
		defaultEdgeFromWithPagination(s, Evidence{}),
		defaultEdgeToWithPagination(s, ControlObjective{}),

		// sub controls can have associated task, narratives, risks, and action plans
		defaultEdgeToWithPagination(s, Task{}),
		defaultEdgeToWithPagination(s, Narrative{}),
		defaultEdgeToWithPagination(s, Risk{}),
		defaultEdgeToWithPagination(s, ActionPlan{}),

		// policies and procedures are used to implement the subcontrol
		defaultEdgeToWithPagination(s, Procedure{}),
		defaultEdgeToWithPagination(s, InternalPolicy{}),

		// owner is the user who is responsible for the subcontrol
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			name:       "control_owner",
			t:          Group.Type,
			comment:    "the user who is responsible for the subcontrol, defaults to the parent control owner if not set",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			name:       "delegate",
			t:          Group.Type,
			comment:    "temporary delegate for the subcontrol, used for temporary control ownership",
		}),
	}
}

// Mixin of the Subcontrol
func (s Subcontrol) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SCL",
		additionalMixins: []ent.Mixin{
			// subcontrols can inherit permissions from the parent control
			NewObjectOwnedMixin(ObjectOwnedMixin{
				FieldNames:               []string{"control_id"},
				WithOrganizationOwner:    true,
				AllowEmptyForSystemAdmin: true, // allow organization owner to be empty
				Ref:                      s.PluralName(),
			}),
		},
	}.getMixins()
}

func (Subcontrol) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the parent control
		index.Fields("control_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
	}
}

// Annotations of the Subcontrol
func (Subcontrol) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Subcontrol
func (Subcontrol) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubcontrolUpdate(),
	}
}

// Policy of the Subcontrol
func (Subcontrol) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.SubcontrolMutation](rule.ControlParent), // if mutation contains control_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.SubcontrolMutation](),
		),
	)
}
