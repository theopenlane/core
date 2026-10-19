package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
)

type AssessmentDomain struct {
	SchemaFuncs

	ent.Schema
}

const SchemaAssessmentDomain = "assessment_domain"

func (AssessmentDomain) Name() string {
	return SchemaAssessmentDomain
}

func (AssessmentDomain) GetType() any {
	return AssessmentDomain.Type
}

func (AssessmentDomain) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAssessmentDomain)
}

func (AssessmentDomain) Fields() []ent.Field {
	return []ent.Field{
		field.String("custom_domain_id").
			NotEmpty(),
	}
}

func (o AssessmentDomain) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			edgeSchema: CustomDomain{},
			field:      "custom_domain_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(CustomDomain{}.Name()),
			},
		}),
	}
}

func (AssessmentDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id").Unique(),
	}
}

func (o AssessmentDomain) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags:       true,
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(o),
		},
	}.getMixins(o)
}

func (AssessmentDomain) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogAssessmentCustomDomainAddon}
}

func (AssessmentDomain) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (AssessmentDomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipMutationCreateInput),
	}
}
