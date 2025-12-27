package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// ActionPlan defines the action plan schema.
type ActionPlan struct {
	SchemaFuncs

	ent.Schema
}

// SchemaActionPlan is the name of the action plan schema.
const SchemaActionPlan = "action_plan"

// Name returns the name of the action plan schema.
func (ActionPlan) Name() string {
	return SchemaActionPlan
}

// GetType returns the type of the action plan schema.
func (ActionPlan) GetType() any {
	return ActionPlan.Type
}

// PluralName returns the plural name of the action plan schema.
func (ActionPlan) PluralName() string {
	return pluralize.NewClient().Plural(SchemaActionPlan)
}

// Fields returns action plan fields.
func (ActionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			NotEmpty().
			Annotations(
				entgql.OrderField("title"),
			).
			Comment("short title describing the action plan"),
		field.Text("description").
			Optional().
			Comment("detailed description of remediation steps and objectives"),
		field.Time("due_date").
			Optional().
			Annotations(
				entgql.OrderField("due_date"),
			).
			Comment("due date of the action plan"),
		field.Time("completed_at").
			Optional().
			Nillable().
			Comment("timestamp when the action plan was completed"),
		field.Enum("priority").
			GoType(enums.Priority("")).
			Annotations(
				entgql.OrderField("PRIORITY"),
			).
			Optional().
			Comment("priority of the action plan"),
		field.Bool("requires_approval").
			Default(false).
			Comment("indicates if the action plan requires explicit approval before closure"),
		field.Bool("blocked").
			Default(false).
			Comment("true when the action plan is currently blocked"),
		field.Text("blocker_reason").
			Optional().
			Comment("context on why the action plan is blocked"),
		field.JSON("metadata", map[string]any{}).
			Optional().
			Comment("additional structured metadata for the action plan"),
		field.JSON("raw_payload", map[string]any{}).
			Optional().
			Comment("raw payload received from the integration for auditing and troubleshooting"),
		field.String("source").
			Annotations(
				entgql.OrderField("source"),
			).
			Optional().
			Comment("source of the action plan"),
	}
}

// Edges of the ActionPlan
func (a ActionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(a, Risk{}),
		defaultEdgeFromWithPagination(a, Control{}),
		defaultEdgeFromWithPagination(a, Program{}),
		defaultEdgeFromWithPagination(a, Finding{}),
		defaultEdgeFromWithPagination(a, Vulnerability{}),
		defaultEdgeFromWithPagination(a, Review{}),
		defaultEdgeFromWithPagination(a, Remediation{}),
		defaultEdgeToWithPagination(a, Task{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Integration{},
			comment:    "integration that generated the action plan",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: a,
			edgeSchema: File{},
			field:      "file_id",
		}),
	}
}

// Mixin of the ActionPlan
func (a ActionPlan) Mixin() []ent.Mixin {
	return mixinConfig{
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			NewDocumentMixin(a),
			newOrgOwnedMixin(a),
			mixin.NewSystemOwnedMixin(),
			newCustomEnumMixin(a),
		}}.getMixins(a)
}

func (ActionPlan) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
		models.CatalogRiskManagementAddon,
		models.CatalogEntityManagementModule,
	}
}

// Annotations of the ActionPlan
func (a ActionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ActionPlan
func (a ActionPlan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ActionPlanMutation](),
		),
	)
}
