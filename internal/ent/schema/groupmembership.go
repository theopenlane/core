package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// GroupMembership holds the schema definition for the GroupMembership entity
type GroupMembership struct {
	SchemaFuncs

	ent.Schema
}

const SchemaGroupMembership = "groupmembership"

func (GroupMembership) Name() string {
	return SchemaGroupMembership
}

func (GroupMembership) GetType() any {
	return GroupMembership.Type
}

func (GroupMembership) PluralName() string {
	return pluralize.NewClient().Plural(SchemaGroupMembership)
}

// Fields of the GroupMembership
func (GroupMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Annotations(
				entgql.OrderField("ROLE"),
			).
			Default(string(enums.RoleMember)),
		field.String("group_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the GroupMembership
func (g GroupMembership) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: Group{},
			required:   true,
			immutable:  true,
			field:      "group_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: User{},
			required:   true,
			immutable:  true,
			field:      "user_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: OrgMembership{},
			immutable:  true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		defaultEdgeToWithPagination(g, Event{}),
	}
}

// Annotations of the GroupMembership
func (GroupMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.MembershipChecks("group"),
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
	return mixinConfig{excludeTags: true}.getMixins()
}

// Interceptors of the GroupMembership
func (GroupMembership) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

// Hooks of the GroupMembership
func (GroupMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookGroupMembers(),
		hooks.HookMembershipSelf("group_memberships"),
	}
}

// Policy of the GroupMembership
func (GroupMembership) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.GroupMembershipMutation](),
		),
	)
}
