package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// VendorRiskScore holds the schema definition for a per-vendor per-question risk assessment
type VendorRiskScore struct {
	SchemaFuncs

	ent.Schema
}

// SchemaVendorRiskScore is the name of the VendorRiskScore schema
const SchemaVendorRiskScore = "vendor_risk_score"

// Name returns the name of the VendorRiskScore schema
func (VendorRiskScore) Name() string {
	return SchemaVendorRiskScore
}

// GetType returns the type of the VendorRiskScore schema
func (VendorRiskScore) GetType() any {
	return VendorRiskScore.Type
}

// PluralName returns the plural name of the VendorRiskScore schema
func (VendorRiskScore) PluralName() string {
	return pluralize.NewClient().Plural(SchemaVendorRiskScore)
}

// Fields of the VendorRiskScore
func (VendorRiskScore) Fields() []ent.Field {
	return []ent.Field{
		field.String("question_key").
			Comment("stable key referencing a VendorScoringQuestionDef; used for grouping across vendors and resolving the current question definition").
			NotEmpty().
			Annotations(
				entgql.OrderField("question_key"),
			),
		field.String("question_name").
			Comment("question text as it existed when this assessment was created; preserved for historical accuracy if the question wording changes later").
			NotEmpty(),
		field.String("question_description").
			Comment("question description captured at assessment time").
			Optional().
			Nillable(),
		field.Enum("question_category").
			Comment("question category captured at assessment time").
			GoType(enums.VendorScoringCategory("")).
			Annotations(
				entgql.OrderField("question_category"),
			),
		field.Enum("answer_type").
			Comment("expected answer format captured at assessment time").
			GoType(enums.VendorScoringAnswerType("")).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			),
		field.Enum("impact").
			Comment("user-assigned impact for this specific vendor using the 5-point TPRM scale (VERY_LOW=1 through CRITICAL=5); the same question may carry different impact across vendors").
			GoType(enums.VendorRiskImpact("")).
			Annotations(
				entgql.OrderField("impact"),
			),
		field.Enum("likelihood").
			Comment("user-assigned likelihood of the risk condition occurring for this vendor using the 5-point TPRM scale (VERY_LOW=0.5 through VERY_HIGH=4)").
			GoType(enums.VendorRiskLikelihood("")).
			Annotations(
				entgql.OrderField("likelihood"),
			),
		field.Float("score").
			Comment("hook-computed risk score: impactNumeric x likelihoodNumeric").
			Default(0).
			Annotations(
				entgql.OrderField("score"),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),
		field.String("answer").
			Comment("factual answer to the question (e.g. 'true', 'false', '48 hours', 'ISO 27001'); retained permanently even if the question text changes, because question_key is the stable reference not the display name").
			Optional().
			Nillable(),
		field.String("notes").
			Comment("optional justification or context for the assigned impact and likelihood").
			Optional().
			Nillable(),
		field.String("vendor_scoring_config_id").
			Comment("the scoring config this assessment belongs to; auto-resolved from org context if not provided").
			Optional(),
		field.String("entity_id").
			Comment("the vendor entity being assessed").
			NotEmpty().
			Annotations(
				entx.CSVRef().FromColumn("VendorRiskScoreEntityName").MatchOn("name"),
			),
		field.String("assessment_response_id").
			Comment("the assessment response this score belongs to; scopes scores to a specific assessment cycle").
			Optional(),
	}
}

// Mixin of the VendorRiskScore
func (v VendorRiskScore) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.VendorRiskScore](v,
				withParents(VendorScoringConfig{}, Entity{}, AssessmentResponse{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(v)
}

// Edges of the VendorRiskScore
func (v VendorRiskScore) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: v,
			name:       "vendor_scoring_config",
			t:          VendorScoringConfig.Type,
			field:      "vendor_scoring_config_id",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: v,
			name:       "entity",
			t:          Entity.Type,
			field:      "entity_id",
			required:   true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: v,
			name:       "assessment_response",
			t:          AssessmentResponse.Type,
			field:      "assessment_response_id",
		}),
	}
}

// Hooks of the VendorRiskScore
func (VendorRiskScore) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookVendorRiskScoreCompute(),
		hooks.HookVendorRiskScoreAggregate(),
	}
}

// Annotations of the VendorRiskScore
func (VendorRiskScore) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
		),
	}
}

// Policy of the VendorRiskScore
func (VendorRiskScore) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				VendorScoringConfig{}.PluralName(),
				Entity{}.PluralName(),
				AssessmentResponse{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
			entfga.CheckEditAccess[*generated.VendorRiskScoreMutation](),
		),
	)
}

// Modules of the VendorRiskScore
func (VendorRiskScore) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
		models.CatalogComplianceModule,
	}
}
