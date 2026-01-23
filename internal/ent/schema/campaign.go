package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// Campaign holds the schema definition for the Campaign entity
type Campaign struct {
	SchemaFuncs

	ent.Schema
}

// SchemaCampaign is the name of the Campaign schema
const SchemaCampaign = "campaign"

// Name returns the name of the Campaign schema
func (Campaign) Name() string {
	return SchemaCampaign
}

// GetType returns the type of the Campaign schema
func (Campaign) GetType() any {
	return Campaign.Type
}

// PluralName returns the plural name of the Campaign schema
func (Campaign) PluralName() string {
	return pluralize.NewClient().Plural(SchemaCampaign)
}

// Fields of the Campaign
func (Campaign) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the campaign").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("the description of the campaign").
			Optional(),
		field.Enum("campaign_type").
			Comment("the type of campaign").
			GoType(enums.CampaignType("")).
			Default(enums.CampaignTypeQuestionnaire.String()).
			Annotations(
				entgql.OrderField("CAMPAIGN_TYPE"),
			),
		field.Enum("status").
			Comment("the status of the campaign").
			GoType(enums.CampaignStatus("")).
			Default(enums.CampaignStatusDraft.String()).
			Annotations(
				entgql.OrderField("STATUS"),
				entx.FieldWorkflowEligible(),
			),
		field.Bool("is_active").
			Comment("whether the campaign is active").
			Default(false).
			Annotations(
				entgql.OrderField("is_active"),
				entx.FieldWorkflowEligible(),
			),
		field.Time("scheduled_at").
			Comment("when the campaign is scheduled to start").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("scheduled_at"),
				entx.FieldWorkflowEligible(),
			),
		field.Time("launched_at").
			Comment("when the campaign was launched").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("launched_at"),
			),
		field.Time("completed_at").
			Comment("when the campaign completed").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("completed_at"),
			),
		field.Time("due_date").
			Comment("when responses are due for the campaign").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("due_date"),
				entx.FieldWorkflowEligible(),
			),
		field.Bool("is_recurring").
			Comment("whether the campaign recurs on a schedule").
			Default(false).
			Annotations(
				entgql.OrderField("is_recurring"),
				entx.FieldWorkflowEligible(),
			),
		field.Enum("recurrence_frequency").
			Comment("the recurrence cadence for the campaign").
			GoType(enums.Frequency("")).
			Optional().
			Annotations(
				entgql.OrderField("recurrence_frequency"),
				entx.FieldWorkflowEligible(),
			),
		field.Int("recurrence_interval").
			Comment("the recurrence interval for the campaign, combined with the recurrence frequency").
			Default(1).
			Optional().
			Annotations(
				entgql.OrderField("recurrence_interval"),
			),
		field.String("recurrence_cron").
			GoType(models.Cron("")).
			Comment("cron schedule to run the campaign in cron 6-field syntax, e.g. 0 0 0 * * *").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Validate(func(s string) error {
				if s == "" {
					return nil
				}

				c := models.Cron(s)

				return c.Validate()
			}).
			Optional().
			Nillable(),
		field.String("recurrence_timezone").
			Comment("timezone used for the recurrence schedule").
			Optional().
			Validate(func(s string) error {
				if s == "" {
					return nil
				}

				_, err := time.LoadLocation(s)

				return err
			}).
			Annotations(
				entgql.OrderField("recurrence_timezone"),
			),
		field.Time("last_run_at").
			Comment("when the campaign was last executed").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("last_run_at"),
			),
		field.Time("next_run_at").
			Comment("when the campaign is scheduled to run next").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("next_run_at"),
			),
		field.Time("recurrence_end_at").
			Comment("when the recurring campaign should stop running").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("recurrence_end_at"),
			),
		field.Int("recipient_count").
			Comment("the number of recipients targeted by the campaign").
			Default(0).
			Optional().
			Annotations(
				entgql.OrderField("recipient_count"),
				entx.FieldWorkflowEligible(),
			),
		field.Int("resend_count").
			Comment("the number of times campaign notifications were resent").
			Default(0).
			Optional().
			Annotations(
				entgql.OrderField("resend_count"),
				entx.FieldWorkflowEligible(),
			),
		field.Time("last_resent_at").
			Comment("when campaign notifications were last resent").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("last_resent_at"),
			),
		field.String("template_id").
			Comment("the template associated with the campaign").
			Optional(),
		field.String("entity_id").
			Comment("the entity associated with the campaign").
			Optional(),
		field.String("assessment_id").
			Comment("the assessment associated with the campaign").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the campaign").
			Optional(),
	}
}

// Mixin of the Campaign
func (c Campaign) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "CMP",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Campaign](c,
				withParents(Organization{}),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			newResponsibilityMixin(c, withInternalOwner()),
			WorkflowApprovalMixin{},
		},
	}.getMixins(c)
}

// Edges of the Campaign
func (c Campaign) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Assessment{},
			field:      "assessment_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Assessment{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Template{},
			field:      "template_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Template{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Entity{},
			field:      "entity_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Entity{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(c, CampaignTarget{}),
		defaultEdgeToWithPagination(c, AssessmentResponse{}),
		defaultEdgeToWithPagination(c, Contact{}),
		defaultEdgeToWithPagination(c, User{}),
		defaultEdgeToWithPagination(c, Group{}),
		defaultEdgeToWithPagination(c, IdentityHolder{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "campaign",
		}),
	}
}

// Indexes of the Campaign
func (Campaign) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("entity_id"),
	}
}

// Modules this schema has access to
func (Campaign) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the Campaign
func (Campaign) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the Campaign
func (Campaign) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			//			entfga.CheckEditAccess[*generated.CampaignMutation](),
		),
	)
}
