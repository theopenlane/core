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
	"github.com/theopenlane/entx/oscalgen"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Risk defines the risk schema.
type Risk struct {
	SchemaFuncs

	ent.Schema
}

// SchemaRisk is the name of the risk schema.
const SchemaRisk = "risk"

// Name returns the name of the risk schema.
func (Risk) Name() string {
	return SchemaRisk
}

// GetType returns the type of the risk schema.
func (Risk) GetType() any {
	return Risk.Type
}

// PluralName returns the plural name of the risk schema.
func (Risk) PluralName() string {
	return pluralize.NewClient().Plural(SchemaRisk)
}

// Fields returns risk fields.
func (Risk) Fields() []ent.Field {
	return []ent.Field{
		field.String("external_id").
			Comment("stable identifier assigned by the source system, used for integration ingest deduplication").
			Optional().
			Annotations(
				entgql.OrderField("external_id"),
				entx.IntegrationMappingField().UpsertKey().LookupKey(),
			),
		field.String("integration_id").
			Comment("integration that surfaced this risk, when sourced via integration ingest").
			Optional().
			Annotations(
				entx.IntegrationMappingField().FromIntegration(),
			),
		field.Time("observed_at").
			Comment("time when this risk was last observed by the source integration").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("observed_at"),
			),
		field.String("external_uuid").
			Comment("stable external UUID for deterministic OSCAL export and round-tripping").
			Optional().
			Nillable().
			Annotations(
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleUUID,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelPOAM, oscalgen.OSCALModelSSP),
					oscalgen.WithOSCALIdentityAnchor(),
				),
			),
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleTitle,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelPOAM, oscalgen.OSCALModelSSP),
				),
				entx.IntegrationMappingField().UpsertKey(),
			).
			Comment("the name of the risk"),
		field.Enum("status").
			GoType(enums.RiskStatus("")).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Optional().
			Comment("status of the risk - identified, mitigated, accepted, closed, transferred, and archived."),
		field.Enum("impact").
			GoType(enums.RiskImpact("")).
			Annotations(
				entgql.OrderField("IMPACT"),
			).
			Optional().
			Comment("impact of the risk -critical, high, medium, low"),
		field.Enum("likelihood").
			GoType(enums.RiskLikelihood("")).
			Default(enums.RiskLikelihoodMid.String()).
			Optional().
			Annotations(
				entgql.OrderField("LIKELIHOOD"),
			).
			Comment("likelihood of the risk occurring; unlikely, likely, highly likely"),
		field.Int("score").
			Optional().
			Annotations(
				entgql.OrderField("score"),
				entx.FieldSearchable(),
			).
			Comment("score of the risk based on impact and likelihood (1-4 unlikely, 5-9 likely, 10-16 highly likely, 17-20 critical)"),
		field.Text("mitigation").
			Optional().
			Comment("mitigation for the risk"),
		field.JSON("mitigation_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment("structured details of the mitigation in JSON format"),
		field.Text("details").
			Optional().
			Annotations(
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleDescription,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelPOAM, oscalgen.OSCALModelSSP),
				),
			).
			Comment("details of the risk"),
		field.JSON("details_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment("structured details of the risk in JSON format"),
		field.Text("business_costs").
			Annotations(
				entgql.OrderField("business_costs"),
			).
			Optional().
			Comment("business costs associated with the risk"),
		field.JSON("business_costs_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment("structured details of the business costs in JSON format"),
		field.String("stakeholder_id").
			Optional().
			Unique().
			Annotations(
				entx.CSVRef().FromColumn("StakeholderGroupName").MatchOn("name"),
			).
			Comment("the id of the group responsible for risk oversight"),
		field.String("delegate_id").
			Optional().
			Unique().
			Annotations(
				entx.CSVRef().FromColumn("RiskDelegateGroupName").MatchOn("name"),
			).
			Comment("the id of the group responsible for risk oversight on behalf of the stakeholder"),

		field.Time("mitigated_at").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("mitigated_at"),
			).
			Comment("the time when the risk was mitigated"),
		field.Bool("review_required").
			Optional().
			Annotations(
				entgql.OrderField("review_required"),
			).
			Default(true).
			Comment("indicates if a periodic review is required for the risk"),
		field.Time("last_reviewed_at").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("last_reviewed_at"),
			).
			Comment("the time when the risk was last reviewed"),
		field.Enum("review_frequency").
			GoType(enums.Frequency("")).
			Default(enums.FrequencyYearly.String()).
			Optional().
			Annotations(
				entgql.OrderField("review_frequency"),
			),
		field.Time("due_date").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("due_date"),
			).
			Comment("the time when the risk is due to be resolved by, based on the sla config but can be manually updated"),
		field.Time("next_review_due_at").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("next_review_due_at"),
			).
			Comment("the time when the next review is due for the risk"),
		field.Int("residual_score").
			Optional().
			Annotations(
				entgql.OrderField("residual_score"),
			).
			Comment("score of the residual risk based on impact and likelihood (1-4 unlikely, 5-9 likely, 10-16 highly likely, 17-20 critical)"),
		field.Enum("risk_decision").
			GoType(enums.RiskDecision("")).
			Default(enums.RiskDecisionNone.String()).
			Optional().
			Annotations(
				entgql.OrderField("risk_decision"),
			).
			Comment("the decision made for the risk - accept, transfer, avoid, mitigate, or none"),
	}
}

// Edges of the Risk
func (r Risk) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Control{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("ControlRefCodes").MatchOn("ref_code"),
				oscalgen.NewOSCALRelationship(
					oscalgen.OSCALRelationshipRoleLinksToControlID,
					oscalgen.WithOSCALRelationshipModels(oscalgen.OSCALModelPOAM, oscalgen.OSCALModelSSP),
				),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Subcontrol{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("SubcontrolRefCodes").MatchOn("ref_code"),
			},
		}),
		defaultEdgeFromWithPagination(r, Procedure{}),
		defaultEdgeFromWithPagination(r, InternalPolicy{}),
		defaultEdgeFromWithPagination(r, Program{}), // risk can be associated to 1:m programs, this allow permission inheritance from the program(s)
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Platform{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("PlatformNames").MatchOn("name"),
				accessmap.EdgeViewCheck(Platform{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: ActionPlan{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("ActionPlanNames").MatchOn("name"),
				accessmap.EdgeViewCheck(ActionPlan{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(r, Task{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Asset{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("AssetNames").MatchOn("name"),
				accessmap.EdgeViewCheck(Asset{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Entity{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("EntityNames").MatchOn("name"),
				accessmap.EdgeViewCheck(Entity{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(r, Scan{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			name:       "stakeholder",
			t:          Group.Type,
			field:      "stakeholder_id",
			comment:    "the group of users who are responsible for risk oversight",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			name:       "delegate",
			t:          Group.Type,
			field:      "delegate_id",
			comment:    "temporary delegates for the risk, used for temporary ownership",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			name:       "comments",
			t:          Note.Type,
			comment:    "conversations related to the risk",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Discussion{},
			comment:    "discussions related to the risk",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
		defaultEdgeFromWithPagination(r, Review{}),
		defaultEdgeFromWithPagination(r, Remediation{}),
	}
}

// Indexes of the Risk
func (Risk) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("external_uuid", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Hooks of the Risk
func (Risk) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"stakeholder": "group",
			}, "stakeholder"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"delegate": "group",
			}, "delegate"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hooks.HookSlateJSON(),
		hooks.HookRisks(),
	}
}

// Mixin of the Risk
func (r Risk) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "RSK",
		additionalMixins: []ent.Mixin{
			// risks inherit permissions from the associated programs, but must have an organization as well
			// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
			// it will also create program parent tuples for the risk when a program is associated to the risk
			newObjectOwnedMixin[generated.Risk](r,
				withParents(
					Program{}, Control{}, Procedure{}, ControlObjective{}, InternalPolicy{}, Subcontrol{}, Platform{}),
				withOrganizationOwner(true),
			),
			// add groups permissions with viewer, editor, and blocked groups
			newGroupPermissionsMixin(),
			newCustomEnumMixin(r),
			newCustomEnumMixin(r, withEnumFieldName("category")),
			newCustomEnumMixin(r, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(r, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(r)
}

func (Risk) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogRiskManagementAddon,
	}
}

// Annotations of the Risk
func (r Risk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
		),
		oscalgen.NewOSCALModel(
			oscalgen.WithOSCALModels(oscalgen.OSCALModelPOAM, oscalgen.OSCALModelSSP),
			oscalgen.WithOSCALAssembly("risk"),
		),
		entx.IntegrationMappingSchema().
			StockPersist().
			Exclude("stakeholder_id", "delegate_id"),
	}
}

// Policy of the Risk
func (r Risk) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			rule.CheckIfCommentOnly(),
			entfga.CheckEditAccess[*generated.RiskMutation](),
		),
	)
}
