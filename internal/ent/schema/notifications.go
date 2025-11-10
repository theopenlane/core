package schema

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

const (
	// Define any constants needed for Notification schema
	titleMaxLen = 256
	bodyMaxLen  = 2000
)

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

// Mixin of the Notification
func (n Notification) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(n),
		},
	}.getMixins(n)
}

// Fields of the Notification
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").
			Comment("the user this notification is for").
			Optional().
			Nillable(),
		field.String("title").
			MaxLen(titleMaxLen).
			NotEmpty().
			Comment("the title of the notification"),
		field.Text("body").
			MaxLen(bodyMaxLen).
			NotEmpty().
			Comment("the body of the notification"),
		field.JSON("data", map[string]interface{}{}).
			Comment("structured payload (IDs, links)").
			Optional(),
		field.Enum("notification_type").
			GoType(enums.NotificationType("")).
			Comment("the type of notification: organization or user"),
		field.String("object_type").
			Comment("the type of the object related to the notification, e.g. taskUpdate, internalPolicyUpdate").
			NotEmpty(),
		field.Time("read_at").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Comment("the time the notification was read"),
		field.Strings("channels").
			Comment("the channels through which the notification was sent: IN_APP, SLACK, EMAIL").
			Optional().
			Validate(func(channels []string) error {
				for _, channel := range channels {
					enumVal := enums.ToNotificationChannel(channel)
					if enumVal == nil || *enumVal == enums.NotificationChannelInvalid {
						return fmt.Errorf("invalid notification channel: %s", channel)
					}
				}

				return nil
			}),
	}
}

// Edges of the Notification
func (n Notification) Edges() []ent.Edge {
	return []ent.Edge{
		// Many notifications → one user (optional)
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: n,
			edgeSchema: User{},
			field:      "user_id",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		// Note: owner edge is automatically created by newOrgOwnedMixin
	}
}

// Indexes of the Notification
func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "read_at", "owner_id"),
	}
}

// Annotations of the Notification
func (Notification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		entgql.Skip(entgql.SkipAll),
	}
}

// Policy of the Notification
func (Notification) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
		),
	)
}

func (Notification) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}
