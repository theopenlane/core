package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
		entfga.MembershipChecks("organization"),
		// Delete groups + program members when orgmembership is deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "Orgmembership",
					Through: "GroupMembership",
				},
				{
					Field:   "Orgmembership",
					Through: "ProgramMembership",
				},
			},
		),
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
		hooks.HookUpdateManagedGroups(),
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
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.OrgMembershipQuery](),
		),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType(&token.OrgInviteToken{}),
			entfga.CheckEditAccess[*generated.OrgMembershipMutation](),
		),
	)
}
