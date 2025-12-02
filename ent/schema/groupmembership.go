package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/interceptors"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/entx/accessmap"
)

// GroupMembership holds the schema definition for the GroupMembership entity
type GroupMembership struct {
	SchemaFuncs

	ent.Schema
}

// SchemaGroupMembership is the name of the GroupMembership schema.
const SchemaGroupMembership = "group_membership"

// Name returns the name of the GroupMembership schema.
func (GroupMembership) Name() string {
	return SchemaGroupMembership
}

// GetType returns the type of the GroupMembership schema.
func (GroupMembership) GetType() any {
	return GroupMembership.Type
}

// PluralName returns the plural name of the GroupMembership schema.
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
			annotations: []schema.Annotation{
				// this is checked via it's own membership check
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: User{},
			required:   true,
			immutable:  true,
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
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

func (GroupMembership) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the GroupMembership
func (g GroupMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.MembershipChecks("group"),
	}
}

// Indexes of the GroupMembership
func (GroupMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "group_id").
			Unique().Annotations(),
	}
}

// Mixin of the GroupMembership
func (GroupMembership) Mixin() []ent.Mixin {
	return mixinConfig{excludeTags: true, excludeSoftDelete: true}.getMixins(GroupMembership{})
}

// Interceptors of the GroupMembership
func (g GroupMembership) Interceptors() []ent.Interceptor {
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
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.GroupMembershipMutation](),
		),
	)
}
