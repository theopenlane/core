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

	"github.com/theopenlane/core/common/models"
)

// VendorScoringConfig holds the schema definition for the VendorScoringConfig entity
type VendorScoringConfig struct {
	SchemaFuncs

	ent.Schema
}

// SchemaVendorScoringConfig is the name of the VendorScoringConfig schema
const SchemaVendorScoringConfig = "vendor_scoring_config"

// Name returns the name of the VendorScoringConfig schema
func (VendorScoringConfig) Name() string {
	return SchemaVendorScoringConfig
}

// GetType returns the type of the VendorScoringConfig schema
func (VendorScoringConfig) GetType() any {
	return VendorScoringConfig.Type
}

// PluralName returns the plural name of the VendorScoringConfig schema
func (VendorScoringConfig) PluralName() string {
	return pluralize.NewClient().Plural(SchemaVendorScoringConfig)
}

// Fields of the VendorScoringConfig
func (VendorScoringConfig) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("questions", models.VendorScoringQuestionsConfig{}).
			Comment("org-custom question overrides and additions; system defaults from models.DefaultVendorScoringQuestions are merged at read time via VendorScoringQuestionsConfig.All()").
			Default(models.VendorScoringQuestionsConfig{}).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipOrderField),
			),
	}
}

// Mixin of the VendorScoringConfig
func (v VendorScoringConfig) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(v),
		},
	}.getMixins(v)
}

// Edges of the VendorScoringConfig
func (v VendorScoringConfig) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(v, VendorRiskScore{}),
	}
}

// Indexes of the VendorScoringConfig
func (VendorScoringConfig) Indexes() []ent.Index {
	return []ent.Index{
		// enforce one scoring config per organization
		index.Fields(ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Hooks of the VendorScoringConfig
func (VendorScoringConfig) Hooks() []ent.Hook {
	return nil
}

// Policy of the VendorScoringConfig
//func (VendorScoringConfig) Policy() ent.Policy {
//	return policy.NewPolicy(
//		policy.WithMutationRules(
//			policy.CheckOrgWriteAccess(),
//		),
//	)
//}

// Annotations of the VendorScoringConfig
func (VendorScoringConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		//		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
		),
	}
}

// Modules of the VendorScoringConfig
func (VendorScoringConfig) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
		models.CatalogComplianceModule,
	}
}
