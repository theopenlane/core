package schema

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"golang.org/x/mod/semver"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/graphapi/directives"
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
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Evidence{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ControlObjective{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Task{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Narrative{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Risk{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ActionPlan{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Procedure{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: InternalPolicy{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			name:       "comments",
			t:          Note.Type,
			comment:    "conversations related to the control",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Discussion{},
			comment:    "discussions related to the control",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
		// owner is the user who is responsible for the control
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "control_owner",
			t:          Group.Type,
			field:      "control_owner_id",
			comment:    "the group of users who are responsible for the control, will be assigned tasks, approval, etc.",
			annotations: []schema.Annotation{
				entgql.OrderField("CONTROL_OWNER_name"),
				accessmap.EdgeViewCheck(Group{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "delegate",
			t:          Group.Type,
			field:      "delegate_id",
			comment:    "temporary delegate for the control, used for temporary control ownership",
			annotations: []schema.Annotation{
				entgql.OrderField("DELEGATE_name"),
				accessmap.EdgeViewCheck(Group{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			name:       "responsible_party",
			t:          Entity.Type,
			field:      "responsible_party_id",
			comment:    "the entity who is responsible for the control implementation when it is a third party",
			annotations: []schema.Annotation{
				entgql.OrderField("RESPONSIBLE_PARTY_name"),
				accessmap.EdgeViewCheck(Entity{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
	}
}

// Hooks of the ControlMixin
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
		hooks.HookSlateJSON(),
	}
}

// Interceptors of the ControlMixin
func (ControlMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorControlFieldSort(),
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
	field.String("external_uuid").
		Comment("stable external UUID for deterministic OSCAL export and round-tripping").
		Optional().
		Nillable().
		Unique(),
	field.String("title").
		Optional().
		Annotations(
			entx.FieldSearchable(),
			entgql.OrderField("title"),
			entx.FieldWebhookPayloadField(),
			directives.ExternalSourceDirectiveAnnotation,
			entx.FieldWorkflowEligible(),
		).
		Comment("human readable title of the control for quick identification"),
	field.Text("description").
		Optional().
		Annotations(
			entx.FieldSearchable(),
			directives.ExternalSourceDirectiveAnnotation,
		).
		Comment("description of what the control is supposed to accomplish"),
	field.JSON("description_json", []any{}).
		Optional().
		Annotations(
			entgql.Type("[Any!]"),
		).
		Comment("structured details of the control in JSON format"),
	field.Strings("aliases").Optional().
		Annotations(
			entx.FieldSearchable(),
		).
		Comment("additional names (ref_codes) for the control"),
	field.String("reference_id").
		Annotations(entx.FieldWorkflowEligible()).
		Optional().
		Comment("internal reference id of the control, can be used for internal tracking"),
	field.String("auditor_reference_id").
		Annotations(entx.FieldWorkflowEligible()).
		Optional().
		Comment("external auditor id of the control, can be used to map to external audit partner mappings"),
	field.String("responsible_party_id").
		Annotations(
			entx.FieldWorkflowEligible(),
			entx.CSVRef().FromColumn("ResponsiblePartyEntityName").MatchOn("name"),
		).
		Optional().
		Comment("the id of the party responsible for the control, usually used when the control is implemented by a third party"),
	field.Enum("status").
		GoType(enums.ControlStatus("")).
		Optional().
		Default(enums.ControlStatusNotImplemented.String()).
		Annotations(
			entgql.OrderField("STATUS"),
			entx.FieldWebhookPayloadField(),
			entx.FieldWorkflowEligible(),
		).
		Comment("status of the control"),
	field.Enum("implementation_status").
		Comment("OSCAL-aligned implementation status of the control").
		GoType(enums.ControlImplementationStatus("")).
		Optional().
		Default(enums.ControlImplementationStatusPlanned.String()),
	field.Text("implementation_description").
		Comment("narrative describing current implementation state for OSCAL export").
		Optional(),
	field.Enum("source").
		GoType(enums.ControlSource("")).
		Optional().
		Annotations(
			entgql.OrderField("SOURCE"),
			directives.ExternalSourceDirectiveAnnotation,
		).
		Default(enums.ControlSourceUserDefined.String()).
		Comment("source of the control, e.g. framework, template, custom, etc."),
	field.String("reference_framework").
		Comment("the reference framework for the control if it came from a standard, empty if not associated with a standard").
		Nillable().
		Annotations(
			entgql.Skip(entgql.SkipMutationUpdateInput),
			entgql.OrderField("REFERENCE_FRAMEWORK"),
			directives.ExternalSourceDirectiveAnnotation,
			entx.FieldWorkflowEligible(),
		).
		Optional(),
	field.String("reference_framework_revision").
		Comment("the reference framework revision for the control if it came from a standard, empty if not associated with a standard, allows for pulling in updates when the standard is updated").
		Nillable().
		Validate(func(s string) error {
			ok := semver.IsValid(s)
			if !ok {
				return fmt.Errorf("%w, invalid semver value", rout.InvalidField("revision"))
			}

			return nil
		}).
		Annotations(
			directives.ExternalSourceDirectiveAnnotation,
			entx.FieldWorkflowEligible(),
		).
		Optional(),
	field.String("category").
		Optional().
		Annotations(
			entx.FieldSearchable(),
			entgql.OrderField("category"),
			entx.FieldWorkflowEligible(),
			directives.ExternalSourceDirectiveAnnotation,
		).
		Comment("category of the control"),
	field.String("category_id").
		Optional().
		Annotations(
			directives.ExternalSourceDirectiveAnnotation,
			entx.FieldWorkflowEligible(),
		).
		Comment("category id of the control"),
	field.String("subcategory").
		Optional().
		Annotations(
			entx.FieldSearchable(),
			entgql.OrderField("subcategory"),
			directives.ExternalSourceDirectiveAnnotation,
			entx.FieldWorkflowEligible(),
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
	field.JSON("testing_procedures", []models.TestingProcedures{}).
		Optional().
		Comment("reference steps to take to test the control"),
	field.JSON("evidence_requests", []models.EvidenceRequests{}).
		Optional().
		Comment("list of common evidence requests for the control"),
	field.String("control_owner_id").
		Optional().
		Nillable().
		Unique().
		Annotations(
			entx.FieldWorkflowEligible(),
			entx.CSVRef().FromColumn("ControlOwnerGroupName").MatchOn("name"),
		).
		Comment("the id of the group that owns the control"),
	field.String("delegate_id").
		Optional().
		Unique().
		Annotations(
			entx.CSVRef().FromColumn("ControlDelegateGroupName").MatchOn("name"),
		).
		Comment("the id of the group that is temporarily delegated to own the control"),
}
