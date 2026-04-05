package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
		field.Enum("scoring_mode").
			Comment("controls how unanswered questions affect the aggregate score: ANSWERED_ONLY sums only answered questions; FULL_QUESTIONNAIRE treats unanswered as maximum risk; MANUAL disables automatic aggregation").
			GoType(enums.VendorScoringMode("")).
			Default(string(enums.VendorScoringModeAnsweredOnly)).
			Annotations(
				entgql.OrderField("scoring_mode"),
			),
		field.JSON("risk_thresholds", models.RiskThresholdsConfig{}).
			Comment("org-custom risk rating threshold overrides; system defaults from models.DefaultRiskThresholds are merged at read time via RiskThresholdsConfig.All()").
			Default(models.RiskThresholdsConfig{}).
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
		edgeToWithPagination(&edgeDefinition{
			fromSchema: v,
			edgeSchema: VendorRiskScore{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(VendorRiskScore{}.Name()),
			},
		}),
	}
}

// Hooks of the VendorScoringConfig
func (VendorScoringConfig) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookVendorScoringConfigKeyGen(),
	}
}

// Policy of the VendorScoringConfig
func (VendorScoringConfig) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the VendorScoringConfig
func (VendorScoringConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
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
