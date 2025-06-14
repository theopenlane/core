package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// ControlMixin implements the control pattern fields for different schemas.
type ControlMixin struct {
	mixin.Schema

	// SchemaType is the schema that implements the SchemaFuncs interface that is using
	// this mixin
	SchemaType any
}

// Fields of the ControlMixin.
func (ControlMixin) Fields() []ent.Field {
	return controlFields
}

// Edges of the ControlMixin.
func (m ControlMixin) Edges() []ent.Edge {
	c := m.SchemaType

	// check if the schema implements the SchemaFuncs interface
	// this happens early to ensure the schema can use the mixin
	if _, ok := c.(SchemaFuncs); !ok {
		panic("ControlMixin must be used with a schema that implements SchemaFuncs")
	}

	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Evidence{}),
		defaultEdgeToWithPagination(c, ControlObjective{}),
		defaultEdgeToWithPagination(c, Task{}),
		defaultEdgeToWithPagination(c, Narrative{}),
		defaultEdgeToWithPagination(c, Risk{}),
		defaultEdgeToWithPagination(c, ActionPlan{}),
		defaultEdgeToWithPagination(c, Procedure{}),
		defaultEdgeFromWithPagination(c, InternalPolicy{}),
		// owner is the user who is responsible for the control
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "control_owner",
			t:          Group.Type,
			field:      "control_owner_id",
			comment:    "the group of users who are responsible for the control, will be assigned tasks, approval, etc.",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "delegate",
			t:          Group.Type,
			field:      "delegate_id",
			comment:    "temporary delegate for the control, used for temporary control ownership",
		}),
	}
}

func (ControlMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"control_owner": "group",
			}, "owner"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"delegate": "group",
			}, "delegate"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}

// Annotations of the Control
func (ControlMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// controlFields are fields use by both Control and Subcontrol schemas
var controlFields = []ent.Field{
	field.Text("description").
		Optional().
		Annotations(
			entx.FieldSearchable(),
		).
		Comment("description of what the control is supposed to accomplish"),
	field.String("reference_id").
		Optional().
		Unique().
		Comment("internal reference id of the control, can be used for internal tracking"),
	field.String("auditor_reference_id").
		Optional().
		Unique().
		Comment("external auditor id of the control, can be used to map to external audit partner mappings"),
	field.Enum("status").
		GoType(enums.ControlStatus("")).
		Optional().
		Default(enums.ControlStatusNotImplemented.String()).
		Annotations(
			entgql.OrderField("STATUS"),
		).
		Comment("status of the control"),
	field.Enum("source").
		GoType(enums.ControlSource("")).
		Optional().
		Annotations(
			entgql.OrderField("SOURCE"),
		).
		Default(enums.ControlSourceUserDefined.String()).
		Comment("source of the control, e.g. framework, template, custom, etc."),
	field.String("reference_framework").
		Comment("the reference framework for the control if it came from a standard, empty if no associated with a standard").
		Nillable().
		Optional(),
	field.Enum("control_type").
		GoType(enums.ControlType("")).
		Default(enums.ControlTypePreventative.String()).
		Annotations(
			entgql.OrderField("CONTROL_TYPE"),
		).
		Optional().
		Comment("type of the control e.g. preventive, detective, corrective, or deterrent."),
	field.String("category").
		Optional().
		Annotations(
			entx.FieldSearchable(),
			entgql.OrderField("category"),
		).
		Comment("category of the control"),
	field.String("category_id").
		Optional().
		Comment("category id of the control"),
	field.String("subcategory").
		Optional().
		Annotations(
			entx.FieldSearchable(),
			entgql.OrderField("subcategory"),
		).
		Comment("subcategory of the control"),
	field.Strings("mapped_categories").
		Optional().
		Annotations(
			entx.FieldSearchable(),
		).
		Comment("mapped categories of the control to other standards"),
	field.JSON("assessment_objectives", []models.AssessmentObjective{}).
		Optional().
		Comment("objectives of the audit assessment for the control"),
	field.JSON("assessment_methods", []models.AssessmentMethod{}).
		Optional().
		Comment("methods used to verify the control implementation during an audit"),
	field.Strings("control_questions").
		Optional().
		Comment("questions to ask to verify the control"),
	field.JSON("implementation_guidance", []models.ImplementationGuidance{}).
		Optional().
		Comment("implementation guidance for the control"),
	field.JSON("example_evidence", []models.ExampleEvidence{}).
		Optional().
		Comment("examples of evidence for the control"),
	field.JSON("references", []models.Reference{}).
		Optional().
		Comment("references for the control"),
	field.String("control_owner_id").
		Optional().
		Unique().
		Comment("the id of the group that owns the control"),
	field.String("delegate_id").
		Optional().
		Unique().
		Comment("the id of the group that is temporarily delegated to own the control"),
}
