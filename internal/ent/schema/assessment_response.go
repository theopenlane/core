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

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// AssessmentResponse stores information about a user's response to an assessment including status, completion, and answers
type AssessmentResponse struct {
	SchemaFuncs
	ent.Schema
}

const SchemaAssessmentResponse = "assessment_response"

func (AssessmentResponse) Name() string { return SchemaAssessmentResponse }
func (AssessmentResponse) GetType() any { return AssessmentResponse.Type }
func (AssessmentResponse) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAssessmentResponse)
}

func (AssessmentResponse) Fields() []ent.Field {
	return []ent.Field{

		field.String("assessment_id").
			Comment("the assessment this response is for").
			Annotations(
				entgql.Skip(entgql.SkipMutationUpdateInput),
			).
			NotEmpty(),

		field.String("email").
			Comment("the email address of the recipient").
			Annotations(
				entx.FieldSearchable(),
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

		field.Enum("status").
			GoType(enums.AssessmentResponseStatus("")).
			Default(enums.AssessmentResponseStatusNotStarted.String()).
			Comment("the current status of the assessment for this user").
			Annotations(
				entgql.OrderField("status"),
			),

		field.Time("assigned_at").
			Comment("when the assessment was assigned to the user").
			Immutable().
			Annotations(
				entgql.OrderField("ASSIGNED_AT"),
			),

		field.Time("started_at").
			Comment("when the user started the assessment").
			Default(time.Now()).
			Annotations(
				entgql.OrderField("STARTED_AT"),
			),
		field.Time("completed_at").
			Comment("when the user completed the assessment").
			Optional().
			Annotations(
				entgql.OrderField("COMPLETED_AT"),
			),
		field.Time("due_date").
			Comment("when the assessment is due").
			Optional().
			Annotations(
				entgql.OrderField("DUE_DATE"),
			),
		field.String("document_data_id").
			Comment("the document containing the user's response data").
			Optional(),
	}
}

func (ar AssessmentResponse) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.AssessmentResponse](ar,
				withParents(Assessment{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(ar)
}

func (ar AssessmentResponse) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: DocumentData{},
			field:      "document_data_id",
			comment:    "the document containing the user's response data",
		}),

		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: Assessment{},
			field:      "assessment_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

func (AssessmentResponse) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the AssessmentResponse
func (AssessmentResponse) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the AssessmentResponse
func (AssessmentResponse) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterQueryResults[generated.AssessmentResponse](),
	}
}

// Indexes of the AssessmentResponse
func (AssessmentResponse) Indexes() []ent.Index {
	return []ent.Index{
		// one response per user per assessment
		index.Fields("assessment_id", "email").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),

		index.Fields("status"),
		index.Fields("due_date"),
		index.Fields("assigned_at"),
		index.Fields("completed_at"),
	}
}

func (AssessmentResponse) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
