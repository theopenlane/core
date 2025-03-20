package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// InternalPolicy defines the policy schema.
type InternalPolicy struct {
	CustomSchema

	ent.Schema
}

const SchemaInternalPolicy = "internal_policy"

func (InternalPolicy) Name() string {
	return SchemaInternalPolicy
}

func (InternalPolicy) GetType() any {
	return InternalPolicy.Type
}

func (InternalPolicy) PluralName() string {
	return pluralize.NewClient().Plural(SchemaInternalPolicy)
}

// Fields returns policy fields.
func (InternalPolicy) Fields() []ent.Field {
	return []ent.Field{} // fields are defined in the mixins
}

// Edges of the InternalPolicy
func (i InternalPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(i, ControlObjective{}),
		defaultEdgeToWithPagination(i, Control{}),
		defaultEdgeToWithPagination(i, Procedure{}),
		defaultEdgeToWithPagination(i, Narrative{}),
		defaultEdgeToWithPagination(i, Task{}),
		defaultEdgeFromWithPagination(i, Program{}),
	}
}

// Mixin of the InternalPolicy
func (InternalPolicy) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "PLC",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			// all policies must be associated to an organization
			NewOrgOwnMixinWithRef("internal_policies"),
			// add group edit permissions to the procedure
			NewGroupPermissionsMixin(false),
			// policies are documents
			DocumentMixin{DocumentType: "policy"},
		},
	}.getMixins()
}

// Annotations of the InternalPolicy
func (InternalPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
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
