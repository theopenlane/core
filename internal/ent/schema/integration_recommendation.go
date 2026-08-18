package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// IntegrationRecommendation stores computed integration recommendations for an organization.
type IntegrationRecommendation struct {
	SchemaFuncs

	ent.Schema
}

const (
	SchemaIntegrationRecommendation = "integration_recommendation"

	maxIntegrationRecommendationWeight = 100
)

func (IntegrationRecommendation) Name() string {
	return SchemaIntegrationRecommendation
}

func (IntegrationRecommendation) GetType() any {
	return IntegrationRecommendation.Type
}

func (IntegrationRecommendation) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegrationRecommendation)
}

func (IntegrationRecommendation) Fields() []ent.Field {
	return []ent.Field{
		field.String("definition_id").
			Comment("the integration definition this recommendation points to").
			NotEmpty(),
		field.Int("weight").
			Comment("computed recommendation weight from 0 to 100").
			Min(0).
			Max(maxIntegrationRecommendationWeight),
		field.String("label").
			Comment("user-facing summary for why this integration was recommended").
			NotEmpty(),
	}
}

func (i IntegrationRecommendation) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(i),
		},
	}.getMixins(i)
}

func (IntegrationRecommendation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "definition_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (IntegrationRecommendation) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
		),
	)
}

func (IntegrationRecommendation) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

func (IntegrationRecommendation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.FGACrudSkip(entx.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}
