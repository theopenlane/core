package schema

import (
	"net/mail"

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

// CampaignTarget holds the schema definition for the CampaignTarget entity
type CampaignTarget struct {
	SchemaFuncs

	ent.Schema
}

// SchemaCampaignTarget is the name of the CampaignTarget schema
const SchemaCampaignTarget = "campaign_target"

// Name returns the name of the CampaignTarget schema
func (CampaignTarget) Name() string {
	return SchemaCampaignTarget
}

// GetType returns the type of the CampaignTarget schema
func (CampaignTarget) GetType() any {
	return CampaignTarget.Type
}

// PluralName returns the plural name of the CampaignTarget schema
func (CampaignTarget) PluralName() string {
	return pluralize.NewClient().Plural(SchemaCampaignTarget)
}

// Fields of the CampaignTarget
func (CampaignTarget) Fields() []ent.Field {
	return []ent.Field{
		field.String("campaign_id").
			Comment("the campaign this target belongs to").
			NotEmpty(),
		field.String("contact_id").
			Comment("the contact associated with the campaign target").
			Optional(),
		field.String("user_id").
			Comment("the user associated with the campaign target").
			Optional(),
		field.String("group_id").
			Comment("the group associated with the campaign target").
			Optional(),
		field.String("email").
			Comment("the email address targeted by the campaign").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("email"),
			).
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("full_name").
			Comment("the name of the campaign target, if known").
			Optional().
			Annotations(
				entgql.OrderField("full_name"),
			),
		field.Enum("status").
			Comment("the delivery or response status for the campaign target").
			GoType(enums.AssessmentResponseStatus("")).
			Default(enums.AssessmentResponseStatusNotStarted.String()).
			Annotations(
				entgql.OrderField("STATUS"),
				entx.FieldWorkflowEligible(),
			),
		field.Time("sent_at").
			Comment("when the campaign target was last sent a request").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("sent_at"),
				entx.FieldWorkflowEligible(),
			),
		field.Time("completed_at").
			Comment("when the campaign target completed the request").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("completed_at"),
				entx.FieldWorkflowEligible(),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the campaign target").
			Optional(),
	}
}

// Mixin of the CampaignTarget
func (c CampaignTarget) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.CampaignTarget](c,
				withParents(Campaign{}),
				withOrganizationOwner(true),
			),
			WorkflowApprovalMixin{},
		},
	}.getMixins(c)
}

// Edges of the CampaignTarget
func (c CampaignTarget) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Campaign{},
			field:      "campaign_id",
			required:   true,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Contact{},
			field:      "contact_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Contact{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: User{},
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(User{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Group{},
			field:      "group_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "campaign_target",
		}),
	}
}

// Indexes of the CampaignTarget
func (CampaignTarget) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("campaign_id", "email").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("status"),
		index.Fields("contact_id"),
		index.Fields("user_id"),
		index.Fields("group_id"),
	}
}

// Modules this schema has access to
func (CampaignTarget) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the CampaignTarget
func (CampaignTarget) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the CampaignTarget
func (CampaignTarget) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Campaign{}.PluralName(),
			}),
			entfga.CheckEditAccess[*generated.CampaignTargetMutation](),
		),
	)
}
