package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
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
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),

		field.Enum("notification_type").
			Comment("the type of notification - organization or user").
			GoType(enums.NotificationType("")).
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.String("object_type").
			Comment("the event type this notification is related to (e.g., task.created, control.updated)").
			NotEmpty().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.String("title").
			Comment("the title of the notification").
			NotEmpty().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.Text("body").
			Comment("the body text of the notification").
			NotEmpty().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.JSON("data", map[string]interface{}{}).
			Comment("structured payload containing IDs, links, and other notification data").
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.String("template_id").
			Comment("optional template used for external channel rendering").
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.Time("read_at").
			Comment("the time the notification was read").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput)),
		field.JSON("channels", []enums.Channel{}).
			Comment("the channels this notification should be sent to (IN_APP, SLACK, EMAIL)").
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
		field.Enum("topic").
			Comment("the topic of the notification (TASK_ASSIGNMENT, APPROVAL, MENTION, EXPORT)").
			GoType(enums.NotificationTopic("")).
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationUpdateInput)),
	}
}

// Hooks of the Notification
func (Notification) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookNotification(),
		hooks.HookNotificationPublish(),
	}
}

// Interceptors of the Notification
func (Notification) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.NotificationQueryFilter(),
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
			immutable:  true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: NotificationTemplate{},
			field:      "template_id",
			immutable:  true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipMutationUpdateInput),
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
		entgql.RelayConnection(),
		// skip schema gen, this is used for subscriptions only
		entgql.Skip(entgql.SkipWhereInput),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()), // generate input types, but we'll skip the create resolver
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
