package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
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
	SchemaFuncs

	ent.Schema
}

// SchemaInternalPolicy is the name of the internal policy schema.
const SchemaInternalPolicy = "internal_policy"

// Name returns the name of the internal policy schema.
func (InternalPolicy) Name() string {
	return SchemaInternalPolicy
}

// GetType returns the type of the internal policy schema.
func (InternalPolicy) GetType() any {
	return InternalPolicy.Type
}

// PluralName returns the plural name of the internal policy schema.
func (InternalPolicy) PluralName() string {
	return pluralize.NewClient().Plural(SchemaInternalPolicy)
}

// Fields returns policy fields.
func (InternalPolicy) Fields() []ent.Field {
	// other fields are defined in the mixins
	return []ent.Field{}
}

// Edges of the InternalPolicy
func (i InternalPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(i, ControlObjective{}),
		defaultEdgeToWithPagination(i, Control{}),
		defaultEdgeToWithPagination(i, Subcontrol{}),
		defaultEdgeToWithPagination(i, Procedure{}),
		defaultEdgeToWithPagination(i, Narrative{}),
		defaultEdgeToWithPagination(i, Task{}),
		defaultEdgeToWithPagination(i, Risk{}),

		defaultEdgeFromWithPagination(i, Program{}),
	}
}

// Mixin of the InternalPolicy
func (i InternalPolicy) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "PLC",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			// all policies must be associated to an organization
			newOrgOwnedMixin(i),
			// add group edit permissions to the procedure
			newGroupPermissionsMixin(withSkipViewPermissions()),
			// policies are documents
			DocumentMixin{DocumentType: "policy"}, // use short name for the document type
		},
	}.getMixins()
}

// Annotations of the InternalPolicy
func (InternalPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("compliance", "policy-management"),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the InternalPolicy
func (InternalPolicy) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.OrgOwnedTuplesHookWithAdmin(),
			ent.OpCreate,
		),
	}
}

// Interceptors of the InternalPolicy
func (InternalPolicy) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// policies are org owned, but we need to ensure the groups are filtered as well
		interceptors.FilterQueryResults[generated.InternalPolicy](),
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
