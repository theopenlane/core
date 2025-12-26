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

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
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
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: ar,
			edgeSchema: Assessment{},
			field:      "assessment_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
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

func (AssessmentResponse) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
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

func (AssessmentResponse) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreateAssessmentResponse(),
		hooks.HookUpdateAssessmentResponse(),
	}
}
