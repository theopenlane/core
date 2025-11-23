package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Notification holds the schema definition for the Notification entity
type Notification struct {
	SchemaFuncs

	ent.Schema
}

const SchemaNotification = "notification"

func (Notification) Name() string {
	return SchemaNotification
}

func (Notification) GetType() any {
	return Notification.Type
}

func (Notification) PluralName() string {
	return pluralize.NewClient().Plural(SchemaNotification)
}

// Fields of the Notification
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").
			Comment("the user this notification is for").
			Optional(),
		field.Enum("notification_type").
			Comment("the type of notification - organization or user").
			GoType(enums.NotificationType("")),
		field.String("object_type").
			Comment("the event type this notification is related to (e.g., task.created, control.updated)").
			NotEmpty(),
		field.String("title").
			Comment("the title of the notification").
			NotEmpty(),
		field.Text("body").
			Comment("the body text of the notification").
			NotEmpty(),
		field.JSON("data", map[string]interface{}{}).
			Comment("structured payload containing IDs, links, and other notification data").
			Optional(),
		field.Time("read_at").
			Comment("the time the notification was read").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.JSON("channels", []enums.Channel{}).
			Comment("the channels this notification should be sent to (IN_APP, SLACK, EMAIL)").
			Optional(),
		field.String("topic").
			Comment("the topic of the notification").
			Optional(),
	}
}

// Hooks of the Notification
func (Notification) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookNotification(),
	}
}

// Mixin of the Notification
func (n Notification) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		excludeSoftDelete:  true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(n),
		},
	}.getMixins(n)
}

// Edges of the Notification
func (n Notification) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: User{},
			field:      "user_id",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

// Indexes of the Notification
func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		// Index for common query: "get all users unread notifications"
		// filters on org, user, and read_at
		index.Fields("user_id", "read_at", "owner_id"),
	}
}

// Modules of the Notification
func (Notification) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the Notification
func (Notification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
		// skip schema gen, this is used for subscriptions only
		entx.SchemaGenSkip(true),
		// skip query gen, this is used for subscriptions only
		entx.QueryGenSkip(true),
		entgql.Skip(entgql.SkipMutationUpdateInput | entgql.SkipWhereInput | entgql.SkipOrderField),
		entgql.Mutations(entgql.MutationCreate()), // only allow creation via mutations, updates are not supported
	}
}

// Policy of the Notification
func (Notification) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			// only allow requests from inside the server to add notifications but not other users
			rule.AllowIfContextAllowRule(),
		),
	)
}
