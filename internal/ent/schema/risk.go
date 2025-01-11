package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
)

// Risk defines the risk schema.
type Risk struct {
	ent.Schema
}

// Fields returns risk fields.
func (Risk) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the name of the risk"),
		field.Text("description").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("description of the risk"),
		field.String("status").
			Optional().
			Comment("status of the risk - mitigated or not, inflight, etc."),
		field.String("risk_type").
			Optional().
			Comment("type of the risk, e.g. strategic, operational, financial, external, etc."),
		field.Text("business_costs").
			Optional().
			Comment("business costs associated with the risk"),
		field.Enum("impact").
			GoType(enums.RiskImpact("")).
			Default(enums.RiskImpactModerate.String()).
			Optional().
			Comment("impact of the risk - high, medium, low"),
		field.Enum("likelihood").
			GoType(enums.RiskLikelihood("")).
			Default(enums.RiskLikelihoodMid.String()).
			Optional().
			Comment("likelihood of the risk occurring; unlikely, likely, highly likely"),
		field.Text("mitigation").
			Optional().
			Comment("mitigation for the risk"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the risk"),
		field.JSON("details", map[string]any{}).
			Optional().
			Comment("json data for the risk document"),
	}
}

// Edges of the Risk
func (Risk) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", Control.Type).
			Ref("risks"),
		edge.From("procedure", Procedure.Type).
			Ref("risks"),
		edge.To("action_plans", ActionPlan.Type),
		edge.From("programs", Program.Type).
			Ref("risks"), // risk can be associated to 1:m programs, this allow permission inheritance from the program(s)
	}
}

// Mixin of the Risk
func (Risk) Mixin() []ent.Mixin {
	return []ent.Mixin{
		NewAuditMixin(),
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		// risks inherit permissions from the associated programs, but must have an organization as well
		// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
		// it will also create program parent tuples for the risk when a program is associated to the risk
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id"},
			WithOrganizationOwner: true,
			Ref:                   "risks",
		}),
		// add groups permissions with viewer, editor, and blocked groups
		NewGroupPermissionsMixin(true),
	}
}

// Annotations of the Risk
func (Risk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entfga.SelfAccessChecks(),
	}
}

// Policy of the Risk
func (Risk) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.RiskQuery](),
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.RiskMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.RiskMutation](),
		),
	)
}
