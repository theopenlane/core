package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

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
		Default(enums.ControlStatusNull.String()).
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
}
