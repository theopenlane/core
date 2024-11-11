package schema

import (
	"context"

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
	"github.com/theopenlane/core/internal/ent/mixin"
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
			Default(string(enums.RoleMember)),
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
	}
}

// Annotations of the ProgramMembership
func (ProgramMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "program",
			IncludeHooks: true,
			IDField:      "ProgramID",
		},
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
	}
}

// // Policy of the ProgramMembership
func (ProgramMembership) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.ProgramMembershipMutationRuleFunc(func(ctx context.Context, m *generated.ProgramMembershipMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.ProgramMembershipQueryRuleFunc(func(ctx context.Context, q *generated.ProgramMembershipQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
