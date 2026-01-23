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

// Remediation defines the remediation schema.
type Remediation struct {
	SchemaFuncs

	ent.Schema
}

// SchemaRemediation is the name of the remediation schema.
const SchemaRemediation = "remediation"

// Name returns the name of the remediation schema.
func (Remediation) Name() string {
	return SchemaRemediation
}

// GetType returns the type of the remediation schema.
func (Remediation) GetType() any {
	return Remediation.Type
}

// PluralName returns the plural name of the remediation schema.
func (Remediation) PluralName() string {
	return pluralize.NewClient().Plural(SchemaRemediation)
}

// Fields returns remediation fields.
func (Remediation) Fields() []ent.Field {
	return []ent.Field{
		field.String("external_id").
			Comment("external identifier from the integration source for the remediation").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("external_id"),
			),
		field.String("external_owner_id").
			Comment("external identifier from the integration source for the remediation").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("external_owner_id"),
			),
		field.String("title").
			Comment("title or short description of the remediation effort").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("title"),
			),
		field.String("state").
			Comment("state of the remediation, such as pending or completed").
			Optional().
			Annotations(
				entgql.OrderField("state"),
			),
		field.String("intent").
			Comment("intent or goal of the remediation effort").
			Optional(),
		field.Text("summary").
			Comment("summary of the remediation approach").
			Optional(),
		field.Text("explanation").
			Comment("detailed explanation of the remediation steps").
			Optional(),
		field.Text("instructions").
			Comment("specific instructions or steps to implement the remediation").
			Optional(),
		field.String("owner_reference").
			Comment("reference to the owner responsible for remediation").
			Optional(),
		field.String("repository_uri").
			Comment("source code repository URI associated with the remediation").
			Optional(),
		field.String("pull_request_uri").
			Comment("pull request URI associated with the remediation").
			Optional(),
		field.String("ticket_reference").
			Comment("reference to a tracking ticket for the remediation").
			Optional(),
		field.Time("due_at").
			Comment("timestamp when the remediation is due").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Time("completed_at").
			Comment("timestamp when the remediation was completed").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Time("pr_generated_at").
			Comment("timestamp when an automated pull request was generated").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Text("error").
			Comment("details about any errors encountered during remediation automation").
			Optional(),
		field.String("source").
			Comment("system that produced the remediation record").
			Optional(),
		field.String("external_uri").
			Comment("link to the remediation in the source system").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("raw metadata payload for the remediation from the source system").
			Optional(),
	}
}

// Edges of the Remediation
func (r Remediation) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Integration{},
			comment:    "integration that produced the remediation",
		}),
		defaultEdgeFromWithPagination(r, Scan{}),
		defaultEdgeToWithPagination(r, Finding{}),
		defaultEdgeToWithPagination(r, Vulnerability{}),
		defaultEdgeToWithPagination(r, ActionPlan{}),
		defaultEdgeToWithPagination(r, Task{}),
		defaultEdgeToWithPagination(r, Control{}),
		defaultEdgeToWithPagination(r, Subcontrol{}),
		defaultEdgeToWithPagination(r, Risk{}),
		defaultEdgeToWithPagination(r, Program{}),
		defaultEdgeToWithPagination(r, Asset{}),
		defaultEdgeToWithPagination(r, Entity{}),
		defaultEdgeToWithPagination(r, Review{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			name:       "comments",
			t:          Note.Type,
			comment:    "discussion threads related to the remediation effort",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: r,
			edgeSchema: File{},
			comment:    "supporting files or evidence for the remediation",
		}),
	}
}

// Mixin of the Remediation
func (r Remediation) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "RMD",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Remediation](r,
				withParents(
					ActionPlan{},
					Program{},
					Control{},
					Subcontrol{},
					Risk{},
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
			newCustomEnumMixin(r, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(r, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(r)
}

// Indexes of the Remediation
func (Remediation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("external_id", "external_owner_id", ownerFieldName).
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
	}
}

// Annotations of the Remediation
func (Remediation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Policy of the Remediation
func (r Remediation) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.RemediationMutation](),
		),
	)
}

func (Remediation) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogVulnerabilityManagementModule,
		models.CatalogRiskManagementAddon,
		models.CatalogComplianceModule,
	}
}
