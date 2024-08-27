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
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
)

// OrgMembership holds the schema definition for the OrgMembership entity
type OrgMembership struct {
	ent.Schema
}

// Fields of the OrgMembership
func (OrgMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Values(string(enums.RoleOwner)). // adds owner to the allowed values
			Default(string(enums.RoleMember)),
		field.String("organization_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the OrgMembership
func (OrgMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
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

// Annotations of the OrgMembership
func (OrgMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType:   "organization",
			IncludeHooks: true,
			IDField:      "OrganizationID",
		},
	}
}

func (OrgMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the OrgMembership
func (OrgMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Hooks of the OrgMembership
func (OrgMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOrgMembers(),
		hooks.HookOrgMembersDelete(),
	}
}

// Interceptors of the OrgMembership
func (OrgMembership) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorOrgMember(),
		interceptors.TraverseOrgMembers(),
	}
}

// Policy of the OrgMembership
func (OrgMembership) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.AllowIfContextHasPrivacyTokenOfType(&token.OrgInviteToken{}),
			privacy.OrgMembershipMutationRuleFunc(func(ctx context.Context, m *generated.OrgMembershipMutation) error {
				return m.CheckAccessForEdit(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.OrgMembershipQueryRuleFunc(func(ctx context.Context, q *generated.OrgMembershipQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
