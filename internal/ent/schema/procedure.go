package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/privacy"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Procedure defines the procedure schema.
type Procedure struct {
	SchemaFuncs

	ent.Schema
}

const SchemaProcedure = "procedure"

func (Procedure) Name() string {
	return SchemaProcedure
}

func (Procedure) GetType() any {
	return Procedure.Type
}

func (Procedure) PluralName() string {
	return pluralize.NewClient().Plural(SchemaProcedure)
}

// Fields returns procedure fields.
func (Procedure) Fields() []ent.Field {
	return []ent.Field{} // fields are defined in the mixin
}

// Edges of the Procedure
func (p Procedure) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(p, Control{}),
		defaultEdgeFromWithPagination(p, InternalPolicy{}),
		defaultEdgeFromWithPagination(p, Program{}),
		defaultEdgeToWithPagination(p, Narrative{}),
		defaultEdgeToWithPagination(p, Risk{}),
		defaultEdgeToWithPagination(p, Task{}),
	}
}

// Mixin of the Procedure
func (p Procedure) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "PRD",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(p),
			// add group edit permissions to the procedure
			newGroupPermissionsMixin(withSkipViewPermissions()),
			// all procedures are documents
			NewDocumentMixin(p),
		},
	}.getMixins()
}

// Annotations of the Procedure
func (Procedure) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Procedure
func (Procedure) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookOrgOwnedTuples(false),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}

// Interceptors of the Procedure
func (Procedure) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// procedures are org owned, but we need to ensure the groups are filtered as well
		interceptors.FilterListQuery(),
	}
}

// Policy of the Procedure
func (Procedure) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ProcedureMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ProcedureMutation](),
		),
	)
}
