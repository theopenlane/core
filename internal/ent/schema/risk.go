package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
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
			Comment("the name of the risk"),
		field.Text("description").
			Optional().
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
		field.JSON("details", map[string]interface{}{}).
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
		edge.To("actionplans", ActionPlan.Type),
		edge.From("program", Program.Type).
			Ref("risks"), // risk can be associated to 1:m programs, this allow permission inheritance from the program(s)
		edge.To("viewers", Group.Type).
			Comment("provides view access to the risk to members of the group"),
		edge.To("editors", Group.Type).
			Comment("provides edit access to the risk to members of the group"),
		edge.To("blocked_groups", Group.Type).
			Comment("groups that are blocked from viewing or editing the risk"),
	}
}

// Mixin of the Risk
func (Risk) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames: []string{"program_id"},
			Required:   false,
		})}
}

// Annotations of the Risk
func (Risk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "risk", // check access to the risk for update/delete
			IncludeHooks: false,
		},
	}
}

// Hooks of the Risk
func (Risk) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"editor_id": "group",
			}, fgax.EditorRelation), // add editor tuples for associated groups
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"blocked_group_id": "group",
			}, fgax.BlockedRelation), // add block tuples for associated groups
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"viewer_id": "group",
			}, fgax.ViewerRelation), // add viewer tuples for associated groups
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}

// Interceptors of the Risk
func (Risk) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

// Policy of the Risk
func (Risk) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.CanCreateObjectsInProgram(), // if mutation contains program_id, check access
			privacy.OnMutationOperation( // if there is no program_id, check access for create in org
				rule.CanCreateObjectsInOrg(),
				ent.OpCreate,
			),
			privacy.RiskMutationRuleFunc(func(ctx context.Context, m *generated.RiskMutation) error {
				return m.CheckAccessForEdit(ctx) // check access for edit
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.RiskQueryRuleFunc(func(ctx context.Context, q *generated.RiskQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
