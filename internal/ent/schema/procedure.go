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
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Procedure defines the procedure schema.
type Procedure struct {
	ent.Schema
}

// Fields returns procedure fields.
func (Procedure) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the procedure").
			Annotations(
				entx.FieldSearchable(),
			).
			NotEmpty(),
		field.Text("description").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("description of the procedure"),
		field.String("status").
			Optional().
			Comment("status of the procedure"),
		field.String("procedure_type").
			Optional().
			Comment("type of the procedure"),
		field.String("version").
			Optional().
			Comment("version of the procedure"),
		field.Text("purpose_and_scope").
			Optional().
			Comment("purpose and scope"),
		field.Text("background").
			Optional().
			Comment("background of the procedure"),
		field.Text("satisfies").
			Optional().
			Comment("which controls are satisfied by the procedure"),
		field.JSON("details", map[string]any{}).
			Optional().
			Comment("json data for the procedure document"),
	}
}

// Edges of the Procedure
func (Procedure) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("controls", Control.Type).
			Ref("procedures"),
		edge.From("internal_policies", InternalPolicy.Type).
			Ref("procedures"),
		edge.To("narratives", Narrative.Type),
		edge.To("risks", Risk.Type),
		edge.To("tasks", Task.Type),
		edge.From("programs", Program.Type).
			Ref("procedures"),
	}
}

// Mixin of the Procedure
func (Procedure) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("procedures"),
		// add group edit permissions to the procedure
		NewGroupPermissionsMixin(false),
	}
}

// Annotations of the Procedure
func (Procedure) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
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
	return []ent.Interceptor{}
}

// Policy of the Procedure
func (Procedure) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.ProcedureQuery](),
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ProcedureMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ProcedureMutation](),
		),
	)
}
