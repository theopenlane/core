package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/entx/accessmap"
)

// Group holds the schema definition for the Group entity
type Group struct {
	SchemaFuncs

	ent.Schema
}

// SchemaGroup is the name of the Group schema.
const SchemaGroup = "group"

// Name returns the name of the Group schema.
func (Group) Name() string {
	return SchemaGroup
}

// GetType returns the type of the Group schema.
func (Group) GetType() any {
	return Group.Type
}

// PluralName returns the plural name of the Group schema.
func (Group) PluralName() string {
	return pluralize.NewClient().Plural(SchemaGroup)
}

// Fields of the Group
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the group - must be unique within the organization").
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			MinLen(minNameLength).
			Validate(validator.SpecialCharValidator).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("the groups description").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Bool("is_managed").
			Comment("whether the group is managed by the system").
			Optional().
			Immutable().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Default(false),
		field.String("gravatar_logo_url").
			Comment("the URL to an auto generated gravatar image for the group").
			Optional().
			Validate(validator.ValidateURL()).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.String("logo_url").
			Comment("the URL to an image uploaded by the customer for the groups avatar image").
			Optional().
			Validate(validator.ValidateURL()).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("display_name").
			Comment("The group's displayed 'friendly' name").
			MaxLen(nameMaxLen).
			Default("").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("display_name"),
			),
		field.String("scim_external_id").
			Comment("the SCIM external ID for the group").
			Optional().
			Nillable(),
		field.String("scim_display_name").
			Comment("the SCIM displayname for the group").
			Optional().
			Nillable(),
		field.Bool("scim_active").
			Comment("whether the SCIM group is marked as active").
			Optional().
			Default(true),
		field.String("scim_group_mailing").
			Comment("the SCIM group mailing list email").
			Optional().
			Nillable(),
	}
}

// Edges of the Group
func (g Group) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			name:       "setting",
			t:          GroupSetting.Type,
			annotations: []schema.Annotation{
				entx.CascadeAnnotationField("Group"),
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		edge.From("users", User.Type).
			Ref("groups").
			// Skip the mutation input for the users edge
			// this should be done via the members edge
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entgql.RelayConnection(),
			).
			Through("members", GroupMembership.Type),
		defaultEdgeToWithPagination(g, Event{}),
		defaultEdgeToWithPagination(g, Integration{}),
		defaultEdgeToWithPagination(g, File{}),
		defaultEdgeToWithPagination(g, Task{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: g,
			edgeSchema: Invite{},
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

// Mixin of the Group
func (g Group) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "GRP",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(g),
			// Add the reverse edges for m:m relationships permissions based on the groups
			newGroupPermissionsEdgesMixin(
				withEdges(Program{}, Risk{}, ControlObjective{}, Narrative{}, ControlImplementation{}, Scan{}, Entity{}),
				withEdgesNoView(Procedure{}, InternalPolicy{}, Control{}, MappedControl{}),
			),
		},
	}.getMixins(g)
}

// Indexes of the Group
func (Group) Indexes() []ent.Index {
	return []ent.Index{
		// We have an organization with many groups, and we want to set the group name to be unique under each organization
		index.Fields("name").
			Edges("owner").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (Group) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Group
func (g Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Delete groups members when groups are deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "Group",
					Through: "GroupMembership",
				},
			},
		),
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the Group
func (g Group) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterQueryResults[generated.Group](),
	}
}

// Hooks of the Group
func (Group) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookGroupAuthz(),
		hooks.HookGroup(),
		hooks.HookManagedGroups(),
	}
}

// Policy of the Group
func (g Group) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.GroupMutation](),
		),
	)
}
