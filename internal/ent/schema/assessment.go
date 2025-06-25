package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
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
			Annotations(
				entgql.OrderField("assessment_type"),
			),
		field.String("template_id").
			Comment("the template id associated with the assessment"),
		field.String("assessment_owner_id").
			Optional().
			Unique().
			Comment("the id of the group that owns the assessment"),
	}
}

func (a Assessment) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
			newGroupPermissionsMixin(),
		},
	}.getMixins()
}

func (a Assessment) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Template{},
			field:      "template_id",
			required:   true,
		}),
		defaultEdgeToWithPagination(a, User{}),
		defaultEdgeToWithPagination(a, AssessmentResponse{}),
		// owner is the group who is responsible for the assessment
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: a,
			name:       "assessment_owner",
			t:          Group.Type,
			field:      "assessment_owner_id",
			comment:    "the group of users who are responsible for the assessment, will be assigned tasks, approval, etc.",
			annotations: []schema.Annotation{
				entgql.OrderField("ASSESSMENT_OWNER_name"),
			},
		}),
	}
}

func (Assessment) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfAssessmentQueryCreatedBy(),
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			rule.AllowIfAssessmentCreatedBy(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the Assessment
func (Assessment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features(entx.ModuleBase),
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
	return []ent.Interceptor{
		interceptors.FilterQueryResults[generated.Assessment](),
	}
}

func (Assessment) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"assessment_owner": "group",
			}, "owner"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}
