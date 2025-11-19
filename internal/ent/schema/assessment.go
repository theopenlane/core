package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/privacy"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// Assessment stores information about an questionnaire filled out
type Assessment struct {
	SchemaFuncs
	ent.Schema
}

const SchemaAssessment = "assessment"

func (Assessment) Name() string       { return SchemaAssessment }
func (Assessment) GetType() any       { return Assessment.Type }
func (Assessment) PluralName() string { return pluralize.NewClient().Plural(SchemaAssessment) }

func (Assessment) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the assessment, e.g. cloud providers, marketing team").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
				entx.FieldSearchable(),
			),

		field.Enum("assessment_type").
			GoType(enums.AssessmentType("")).
			Default(enums.AssessmentTypeInternal.String()).
			Immutable().
			Annotations(
				entgql.OrderField("assessment_type"),
			),

		field.String("template_id").
			Optional().
			Comment("the template id associated with this assessment. You can either provide this alone or provide both the jsonconfig and uischema"),

		field.JSON("jsonconfig", map[string]any{}).
			Comment("the jsonschema object of the questionnaire. If not provided it will be inherited from the template.").
			Optional().
			Annotations(
				entx.FieldJSONPathSearchable("$id"),
			),

		field.JSON("uischema", map[string]any{}).
			Comment("the uischema for the template to render in the UI. If not provided, it will be inherited from the template").
			Optional().
			Annotations(),

		field.Int64("response_due_duration").
			Optional().
			Comment("the duration in seconds that the user has to complete the assessment response, defaults to 7 days").
			Annotations(
				entgql.OrderField("response_due_duration"),
			),
	}
}

func (a Assessment) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
			newGroupPermissionsMixin(),
		},
	}.getMixins(a)
}

func (a Assessment) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Template{},
			field:      "template_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Template{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(a, AssessmentResponse{}),
	}
}

func (Assessment) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithOnMutationRules(
			ent.OpDelete|ent.OpDeleteOne,
			policy.CheckOrgWriteAccess(),
		),
		policy.WithMutationRules(
			privacy.AlwaysAllowRule(),
		),
	)
}

// Annotations of the Assessment
func (Assessment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Indexes of the Assessment
func (Assessment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Interceptors of the Assessment
func (Assessment) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

func (Assessment) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookQuestionnaireAssessment(),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"assessment_owner": "group",
			}, "owner"),
			ent.OpCreate|ent.OpUpdateOne,
		),
	}
}

func (Assessment) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
