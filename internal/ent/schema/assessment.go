package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
)

// Assessment stores information about a discovered asset such as technology, domain, or device.
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
		field.String("questionnaire_id").
			Comment("the questionnaire template id associated with the assessment").
			Optional(),
	}
}

func (a Assessment) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
		},
	}.getMixins()
}

func (a Assessment) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(a, User{}),
		defaultEdgeToWithPagination(a, AssessmentResponse{}),
	}
}

func (Assessment) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the Assessment
func (Assessment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features(entx.ModuleBase),
	}
}

// Indexes of the Assessment
func (Assessment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}
