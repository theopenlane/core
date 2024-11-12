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

	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Group holds the schema definition for the Group entity
type Group struct {
	ent.Schema
}

// Mixin of the Group
func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("groups"),
	}
}

// Fields of the Group
func (Group) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the group - must be unique within the organization").
			NotEmpty().
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
		field.String("gravatar_logo_url").
			Comment("the URL to an auto generated gravatar image for the group").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("logo_url").
			Comment("the URL to an image uploaded by the customer for the groups avatar image").
			Optional().
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
	}
}

// Edges of the Group
func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("setting", GroupSetting.Type).
			Required().
			Unique().
			Annotations(
				entx.CascadeAnnotationField("Group"),
			),
		edge.From("users", User.Type).
			Ref("groups").
			Through("members", GroupMembership.Type),
		edge.To("events", Event.Type),
		edge.To("integrations", Integration.Type),
		edge.To("files", File.Type),
		edge.To("tasks", Task.Type),
	}
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

// Annotations of the Group
func (Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		// Delete groups members when groups are deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "Group",
					Through: "GroupMembership",
				},
			},
		),
		entfga.Annotations{
			ObjectType:   "group",
			IncludeHooks: false,
		},
	}
}

// Policy of the group
func (Group) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			privacy.OnMutationOperation(
				rule.CanCreateObjectsInOrg(),
				ent.OpCreate,
			),
			privacy.OnMutationOperation(
				privacy.GroupMutationRuleFunc(func(ctx context.Context, m *generated.GroupMutation) error {
					return m.CheckAccessForEdit(ctx)
				}),
				ent.OpUpdate|ent.OpUpdateOne,
			),
			privacy.OnMutationOperation(
				privacy.GroupMutationRuleFunc(func(ctx context.Context, m *generated.GroupMutation) error {
					return m.CheckAccessForDelete(ctx)
				}),
				ent.OpDelete|ent.OpDeleteOne,
			),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.GroupQueryRuleFunc(func(ctx context.Context, q *generated.GroupQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}

// Interceptors of the Group
func (Group) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorGroup(),
	}
}

// Hooks of the Group
func (Group) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookGroupAuthz(),
		hooks.HookGroup(),
	}
}
