package schema

import (
	"net/mail"
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
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// AssessmentResponse stores information about a user's response to an assessment including status, completion, and answers
type AssessmentResponse struct {
	SchemaFuncs
	ent.Schema
}

// SchemaAssessmentResponse is the stable schema name for assessment responses.
const SchemaAssessmentResponse = "assessment_response"

// Name returns the schema name for AssessmentResponse.
func (AssessmentResponse) Name() string { return SchemaAssessmentResponse }

// GetType returns the schema's Type reference.
func (AssessmentResponse) GetType() any { return AssessmentResponse.Type }

// PluralName returns the plural schema name for AssessmentResponse.
func (AssessmentResponse) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAssessmentResponse)
}

// Fields defines the AssessmentResponse fields.
func (AssessmentResponse) Fields() []ent.Field {
	return []ent.Field{

		field.String("assessment_id").
			Comment("the assessment this response is for").
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput),
			).
			NotEmpty(),
		field.Bool("is_test").
			Comment("whether this assessment response is for a test send").
			Default(false).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.String("campaign_id").
			Comment("the campaign this response is associated with").
			Optional(),
		field.String("identity_holder_id").
			Comment("the identity holder record for the recipient").
			Optional().
			Annotations(
				entx.CSVRef().FromColumn("AssessmentIdentityHolderEmail").MatchOn("email"),
			),
		field.String("entity_id").
			Comment("the entity associated with this assessment response").
			Optional().
			Annotations(
				entx.CSVRef().FromColumn("AssessmentResponseEntityName").MatchOn("name"),
			),

		field.String("email").
			Comment("the email address of the recipient").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("email"),
			).
			Immutable().
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),

		field.Int("send_attempts").
			Comment("the number of attempts made to perform email send to the recipient about this assessment, maximum of 5").
			Annotations(
				entgql.OrderField("send_attempts"),
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Default(1),
		field.Time("email_delivered_at").
			Comment("when the assessment email was delivered to the recipient").
			Optional().
			Annotations(
				entgql.OrderField("email_delivered_at"),
			),
		field.Time("email_opened_at").
			Comment("when the assessment email was opened by the recipient").
			Optional().
			Annotations(
				entgql.OrderField("email_opened_at"),
			),
		field.Time("email_clicked_at").
			Comment("when a link in the assessment email was clicked by the recipient").
			Optional().
			Annotations(
				entgql.OrderField("email_clicked_at"),
			),
		field.Int("email_open_count").
			Comment("the number of times the assessment email was opened").
			Default(0).
			Optional().
			Annotations(
				entgql.OrderField("email_open_count"),
			),
		field.Int("email_click_count").
			Comment("the number of link clicks for the assessment email").
			Default(0).
			Optional().
			Annotations(
				entgql.OrderField("email_click_count"),
			),
		field.Time("last_email_event_at").
			Comment("the most recent email event timestamp for this assessment response").
			Optional().
			Annotations(
				entgql.OrderField("last_email_event_at"),
			),
		field.JSON("email_metadata", map[string]any{}).
			Comment("additional metadata about email delivery events").
			Optional(),

		field.Enum("status").
			GoType(enums.AssessmentResponseStatus("")).
			Default(enums.AssessmentResponseStatusSent.String()).
			Comment("the current status of the assessment for this user").
			Annotations(
				entgql.OrderField("status"),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),

		field.Time("assigned_at").
			Comment("when the assessment was assigned to the user").
			Immutable().
			Default(time.Now).
			Annotations(
				entgql.OrderField("assigned_at"),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),

		field.Time("started_at").
			Comment("when the user started the assessment").
			Default(time.Now()).
			Annotations(
				entgql.OrderField("started_at"),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),
		field.Time("completed_at").
			Comment("when the user completed the assessment").
			Optional().
			Annotations(
				entgql.OrderField("completed_at"),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),
		field.Time("due_date").
			Comment("when the assessment response is due").
			Optional().
			Annotations(
				entgql.OrderField("due_date"),
			),
		field.String("document_data_id").
			Optional().
			Annotations(
				entgql.Skip(^entgql.SkipType),
			).
			Comment("the document containing the user's response data"),

		field.Bool("is_draft").
			Default(false).
			Comment("is this a draft response? can the user resume from where they left?").
			Annotations(
				entgql.OrderField("is_draft"),
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			),
	}
}

// Mixin configures shared mixins for AssessmentResponse.
func (ar AssessmentResponse) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.AssessmentResponse](ar,
				withParents(Assessment{}, Campaign{}, IdentityHolder{}, Entity{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(ar)
}

// Edges defines the AssessmentResponse relationships.
func (ar AssessmentResponse) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: Assessment{},
			field:      "assessment_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: Campaign{},
			field:      "campaign_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Campaign{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: IdentityHolder{},
			field:      "identity_holder_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(IdentityHolder{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: Entity{},
			field:      "entity_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Entity{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: DocumentData{},
			field:      "document_data_id",
			name:       "document_data",
			required:   false,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Policy configures authorization for AssessmentResponse operations.
func (AssessmentResponse) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Campaign{}.PluralName(),
				Entity{}.PluralName(),
			}),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the AssessmentResponse
func (AssessmentResponse) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entgql.Skip(
			entgql.SkipMutationUpdateInput,
		),
	}
}

// Interceptors of the AssessmentResponse
func (AssessmentResponse) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Indexes of the AssessmentResponse
func (AssessmentResponse) Indexes() []ent.Index {
	return []ent.Index{
		// one response per user per assessment when not tied to a campaign
		index.Fields("assessment_id", "email", "is_test").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL AND campaign_id IS NULL")),

		// one response per campaign + assessment + recipient + test flag
		index.Fields("campaign_id", "assessment_id", "email", "is_test").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL AND campaign_id IS NOT NULL")),

		index.Fields("campaign_id"),
		index.Fields("identity_holder_id"),
		index.Fields("entity_id"),
		index.Fields("status"),
		index.Fields("due_date"),
		index.Fields("assigned_at"),
		index.Fields("completed_at"),
	}
}

// Modules declares the modules required for AssessmentResponse.
func (AssessmentResponse) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Hooks configures the AssessmentResponse hooks.
func (AssessmentResponse) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreateAssessmentResponse(),
		hooks.HookUpdateAssessmentResponse(),
	}
}
