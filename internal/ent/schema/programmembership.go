package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// ProgramMembership holds the schema definition for the ProgramMembership entity
type ProgramMembership struct {
	ent.Schema
}

// Fields of the ProgramMembership
func (ProgramMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Default(string(enums.RoleMember)).
			Annotations(
				entgql.OrderField("ROLE"),
			),
		field.String("program_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the ProgramMembership
func (ProgramMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("program", Program.Type).
			Field("program_id").
			Required().
			Unique().
			Immutable(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique().
			Immutable(),
		edge.To("orgmembership", OrgMembership.Type).
			Immutable().
			Unique().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
	}
}

// Annotations of the ProgramMembership
func (ProgramMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.MembershipChecks("program"),
	}
}

// Indexes of the ProgramMembership
func (ProgramMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "program_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the ProgramMembership
func (ProgramMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Hooks of the ProgramMembership
func (ProgramMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookProgramMembers(),
		hooks.HookMembershipSelf("program_memberships"),
	}
}

// Interceptors of the ProgramMembership
func (ProgramMembership) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

// // Policy of the ProgramMembership
func (ProgramMembership) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.ProgramMembershipMutation](),
		),
	)
}
