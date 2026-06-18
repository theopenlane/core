package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// OrgMembership holds the schema definition for the OrgMembership entity
type OrgMembership struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrgMembership = "org_membership"

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
			Values(enums.RoleOwner.String(), enums.RoleSuperAdmin.String(), enums.RoleAuditor.String()).
			Annotations(
				entgql.OrderField("ROLE"),
			).
			Default(enums.RoleMember.String()),
		field.String("organization_id").Immutable(),
		field.String("user_id").Immutable(),
		field.Bool("sso_exempt").
			Comment("member is exempt from the SSO login redirect for this organization; TFA enforcement still applies. Who may set this is gated by the org membership mutation policy").
			Default(false).
			Optional(),
		field.String("sso_exempt_reason").
			Comment("reason the member was granted an SSO exemption").
			Optional().
			Nillable(),
		field.String("sso_exempt_granted_by").
			Comment("id of the user that granted the SSO exemption; stamped server-side, not settable via the API").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			).
			Optional().
			Nillable(),
		field.Time("sso_exempt_granted_at").
			Comment("when the SSO exemption was granted; stamped server-side, not settable via the API").
			GoType(models.DateTime{}).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			).
			Optional().
			Nillable(),
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
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: o,
			edgeSchema: User{},
			required:   true,
			immutable:  true,
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
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
		entx.FGACrudParent(Organization{}.Name()),
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
	return mixinConfig{excludeTags: true, excludeSoftDelete: true}.getMixins(OrgMembership{})
}

// Hooks of the OrgMembership
func (OrgMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookBlockOwnerRoleChange(),
		hooks.HookUpdateManagedGroups(),
		hooks.HookOrgMembers(),
		hooks.HookMembershipSelf("org_memberships"),
		hooks.HookOrgMembersDelete(),
		hooks.HookSSOExemptionAttribution(),
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
			rule.AllowQueryIfSystemAdmin(),
		),
		policy.WithOnMutationRules(
			ent.OpDelete|ent.OpDeleteOne,
			rule.AllowSelfOrgMembershipDelete(),
		),
		policy.WithOnMutationRules(
			ent.OpUpdate|ent.OpUpdateOne,
			rule.AllowOrgMemberRoleUpdate(),
		),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.OrgInviteToken](),
			entfga.CheckEditAccess[*generated.OrgMembershipMutation](),
		),
	)
}

func (OrgMembership) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}
