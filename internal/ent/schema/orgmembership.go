package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
)

// OrgMembership holds the schema definition for the OrgMembership entity
type OrgMembership struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrgMembership = "orgmembership"

func (OrgMembership) Name() string {
	return SchemaOrgMembership
}

func (OrgMembership) GetType() any {
	return OrgMembership.Type
}

func (OrgMembership) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrgMembership)
}

// Fields of the OrgMembership
func (OrgMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Annotations(
				entgql.OrderField("ROLE"),
			).
			Values(string(enums.RoleOwner)). // adds owner to the allowed values
			Default(string(enums.RoleMember)),
		field.String("organization_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the OrgMembership
func (o OrgMembership) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			edgeSchema: Organization{},
			required:   true,
			immutable:  true,
			field:      "organization_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			edgeSchema: User{},
			required:   true,
			immutable:  true,
			field:      "user_id",
		}),
		defaultEdgeToWithPagination(o, Event{}),
	}
}

// Annotations of the OrgMembership
func (OrgMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.MembershipChecks("organization"),
		// Delete groups + program members when orgmembership is deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "User", // use the user field because the orgmembership is deleted
					Through: "GroupMembership",
				},
				{
					Field:   "User", // use the user field because the orgmembership is deleted
					Through: "ProgramMembership",
				},
			},
		),
	}
}

func (OrgMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").
			Unique().Annotations(),
	}
}

// Mixin of the OrgMembership
func (OrgMembership) Mixin() []ent.Mixin {
	return mixinConfig{excludeTags: true, excludeSoftDelete: true}.getMixins()
}

// Hooks of the OrgMembership
func (OrgMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookUpdateManagedGroups(),
		hooks.HookOrgMembers(),
		hooks.HookMembershipSelf("org_memberships"),
		hooks.HookOrgMembersDelete(),
		// hooks.HookOrgMembersFGACleanup(),
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
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithOnMutationRules(
			ent.OpDelete|ent.OpDeleteOne,
			rule.AllowSelfOrgMembershipDelete(),
		),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OrgInviteToken](),
			entfga.CheckEditAccess[*generated.OrgMembershipMutation](),
		),
	)
}
