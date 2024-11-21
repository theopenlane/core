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

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// InternalPolicy defines the policy schema.
type InternalPolicy struct {
	ent.Schema
}

// Fields returns policy fields.
func (InternalPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the policy").
			NotEmpty(),
		field.Text("description").
			Optional().
			Comment("description of the policy"),
		field.String("status").
			Optional().
			Comment("status of the policy"),
		field.String("policy_type").
			Optional().
			Comment("type of the policy"),
		field.String("version").
			Optional().
			Comment("version of the policy"),
		field.Text("purpose_and_scope").
			Optional().
			Comment("purpose and scope"),
		field.Text("background").
			Optional().
			Comment("background of the policy"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data for the policy document"),
	}
}

// Edges of the InternalPolicy
func (InternalPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controlobjectives", ControlObjective.Type),
		edge.To("controls", Control.Type),
		edge.To("procedures", Procedure.Type),
		edge.To("narratives", Narrative.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("policies"),
		edge.To("editors", Group.Type).
			Comment("provides edit access to the policy to members of the group"),
		edge.To("blocked_groups", Group.Type).
			Comment("groups that are blocked from viewing or editing the policy"),
	}
}

// Mixin of the InternalPolicy
func (InternalPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		// all policies must be associated to an organization
		NewOrgOwnMixinWithRef("internalpolicies"),
	}
}

// Annotations of the InternalPolicy
func (InternalPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "internalpolicy",
			IncludeHooks: false,
		},
	}
}

// Hooks of the InternalPolicy
func (InternalPolicy) Hooks() []ent.Hook {
	return []ent.Hook{
		// add org owner tuples for associated organizations
		hook.On(
			hooks.HookOrgOwnedTuples(false),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		// add editor tuples for associated groups
		hook.On(
			hooks.HookEditorTuples(map[string]string{
				"editor_id": "group",
			}), // add editor tuples for associated groups
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookBlockedTuples(map[string]string{
				"blocked_group_id": "group",
			}), // add block tuples for associated groups
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}

// Interceptors of the InternalPolicy
func (InternalPolicy) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

// Policy of the InternalPolicy
func (InternalPolicy) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.OnMutationOperation(
				rule.CanCreateObjectsInOrg(),
				ent.OpCreate,
			),
			privacy.OnMutationOperation(
				privacy.InternalPolicyMutationRuleFunc(func(ctx context.Context, m *generated.InternalPolicyMutation) error {
					return m.CheckAccessForEdit(ctx)
				}),
				ent.OpUpdate|ent.OpUpdateOne|ent.OpUpdate,
			),
			privacy.OnMutationOperation(
				privacy.InternalPolicyMutationRuleFunc(func(ctx context.Context, m *generated.InternalPolicyMutation) error {
					return m.CheckAccessForDelete(ctx)
				}),
				ent.OpDelete|ent.OpDeleteOne,
			),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.InternalPolicyQueryRuleFunc(func(ctx context.Context, q *generated.InternalPolicyQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
