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

	emixin "github.com/datumforge/entx/mixin"
	"github.com/datumforge/fgax/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/pkg/enums"
)

// GroupMembership holds the schema definition for the GroupMembership entity
type GroupMembership struct {
	ent.Schema
}

// Fields of the GroupMembership
func (GroupMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Default(string(enums.RoleMember)),
		field.String("group_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the GroupMembership
func (GroupMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("group", Group.Type).
			Field("group_id").
			Required().
			Unique().
			Immutable(),
		edge.To("user", User.Type).
			Field("user_id").
			Required().
			Unique().
			Immutable(),
		edge.To("events", Event.Type),
	}
}

// Annotations of the GroupMembership
func (GroupMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "group",
			IncludeHooks: true,
			IDField:      "GroupID",
		},
	}
}

// Indexes of the GroupMembership
func (GroupMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "group_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the GroupMembership
func (GroupMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Hooks of the GroupMembership
func (GroupMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookGroupMembers(),
	}
}

// Policy of the GroupMembership
func (GroupMembership) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.GroupMembershipMutationRuleFunc(func(ctx context.Context, m *generated.GroupMembershipMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.GroupMembershipQueryRuleFunc(func(ctx context.Context, q *generated.GroupMembershipQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
