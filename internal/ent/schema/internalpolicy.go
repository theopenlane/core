package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
			Annotations(
				entx.FieldSearchable(),
			).
			NotEmpty(),
		field.Text("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Optional().
			Comment("description of the policy"),
		field.String("status").
			Optional().
			Comment("status of the policy"),
		field.Time("review_due").
			Comment("the date the policy should be reviewed, defaults to a year from creation date").
			Default(time.Now().AddDate(1, 0, 0)).
			Optional(),
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
		field.JSON("details", map[string]any{}).
			Optional().
			Comment("json data for the policy document"),
	}
}

// Edges of the InternalPolicy
func (InternalPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("control_objectives", ControlObjective.Type),
		edge.To("controls", Control.Type),
		edge.To("procedures", Procedure.Type),
		edge.To("narratives", Narrative.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("internal_policies"),
	}
}

// Mixin of the InternalPolicy
func (InternalPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.NewIDMixinWithPrefixedID("PLC"),
		emixin.TagMixin{},
		// all policies must be associated to an organization
		NewOrgOwnMixinWithRef("internal_policies"),
		// add group edit permissions to the procedure
		NewGroupPermissionsMixin(false),
	}
}

// Annotations of the InternalPolicy
func (InternalPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the InternalPolicy
func (InternalPolicy) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookOrgOwnedTuples(false),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}

// Interceptors of the InternalPolicy
func (InternalPolicy) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// procedures are org owned, but we need to ensure the groups are filtered as well
		interceptors.FilterListQuery(),
	}
}

// Policy of the InternalPolicy
func (InternalPolicy) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.InternalPolicyMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.InternalPolicyMutation](),
		),
	)
}
