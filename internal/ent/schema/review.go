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
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// Review defines the review schema.
type Review struct {
	SchemaFuncs

	ent.Schema
}

// SchemaReview is the name of the review schema.
const SchemaReview = "review"

// Name returns the name of the review schema.
func (Review) Name() string {
	return SchemaReview
}

// GetType returns the type of the review schema.
func (Review) GetType() any {
	return Review.Type
}

// PluralName returns the plural name of the review schema.
func (Review) PluralName() string {
	return pluralize.NewClient().Plural(SchemaReview)
}

// Fields returns review fields.
func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.String("external_id").
			Comment("external identifier from the integration source for the review").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("external_id"),
			),
		field.String("external_owner_id").
			Comment("external identifier from the integration source for the review").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("external_owner_id"),
			),
		field.String("title").
			Comment("title of the review").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("title"),
			),
		field.String("state").
			Comment("state of the review").
			Optional().
			Annotations(
				entgql.OrderField("state"),
			),
		field.String("category").
			Comment("category for the review record").
			Optional(),
		field.String("classification").
			Comment("classification or sensitivity of the review record").
			Optional(),
		field.Text("summary").
			Comment("summary text for the review").
			Optional(),
		field.Text("details").
			Comment("detailed notes captured during the review").
			Optional(),
		field.String("reporter").
			Comment("person or system that created the review").
			Optional(),
		field.Bool("approved").
			Comment("true when the review has been approved").
			Default(false).
			Optional(),
		field.Time("reviewed_at").
			Comment("timestamp when the review was completed").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Time("reported_at").
			Comment("timestamp when the review was reported or opened").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Time("approved_at").
			Comment("timestamp when the review was approved").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.String("reviewer_id").
			Comment("identifier for the user primarily responsible for the review").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.String("source").
			Comment("system that produced the review record").
			Optional(),
		field.String("external_uri").
			Comment("link to the review in the source system").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("raw metadata payload for the review from the source system").
			Optional(),
		field.JSON("raw_payload", map[string]any{}).
			Comment("raw payload received from the integration for auditing and troubleshooting").
			Optional(),
	}
}

// Edges of the Review
func (r Review) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Integration{},
			comment:    "integration that produced the review",
		}),
		defaultEdgeToWithPagination(r, Finding{}),
		defaultEdgeToWithPagination(r, Vulnerability{}),
		defaultEdgeToWithPagination(r, ActionPlan{}),
		defaultEdgeToWithPagination(r, Remediation{}),
		defaultEdgeToWithPagination(r, Control{}),
		defaultEdgeToWithPagination(r, Subcontrol{}),
		defaultEdgeToWithPagination(r, Risk{}),
		defaultEdgeToWithPagination(r, Program{}),
		defaultEdgeToWithPagination(r, Asset{}),
		defaultEdgeToWithPagination(r, Entity{}),
		defaultEdgeToWithPagination(r, Task{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			name:       "reviewer",
			t:          User.Type,
			field:      "reviewer_id",
			comment:    "primary reviewer responsible for the record",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			name:       "comments",
			t:          Note.Type,
			comment:    "discussion threads captured during the review",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: File{},
			comment:    "supporting files or evidence for the review",
		}),
	}
}

// Mixin of the Review
func (r Review) Mixin() []ent.Mixin {
	return mixinConfig{
		// prefix: "RVW",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Review](r,
				withParents(
					Program{},
					Control{},
					Subcontrol{},
					Risk{},
					ActionPlan{},
					Finding{},
					Vulnerability{},
					Asset{},
					Entity{},
					Task{},
				),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
		},
	}.getMixins(r)
}

// Indexes of the Review
func (Review) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("external_id", "external_owner_id", ownerFieldName).
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
	}
}

// Annotations of the Review
func (Review) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Policy of the Review
func (r Review) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ReviewMutation](),
		),
	)
}

func (Review) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogVulnerabilityManagementModule,
	}
}
