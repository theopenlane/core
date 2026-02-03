package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/accessmap"
)

// NotificationPreference holds the schema definition for notification preferences.
type NotificationPreference struct {
	SchemaFuncs

	ent.Schema
}

// SchemaNotificationPreference is the name of the NotificationPreference schema.
const SchemaNotificationPreference = "notification_preference"

// Name returns the name of the NotificationPreference schema.
func (NotificationPreference) Name() string {
	return SchemaNotificationPreference
}

// GetType returns the type of the NotificationPreference schema.
func (NotificationPreference) GetType() any {
	return NotificationPreference.Type
}

// PluralName returns the plural name of the NotificationPreference schema.
func (NotificationPreference) PluralName() string {
	return pluralize.NewClient().Plural(SchemaNotificationPreference)
}

// Fields of the NotificationPreference.
func (NotificationPreference) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").
			Comment("the user this preference applies to").
			Immutable(),
		field.Enum("channel").
			Comment("the channel this preference applies to").
			GoType(enums.Channel("")).
			Annotations(
				entgql.OrderField("CHANNEL"),
			),
		field.Enum("status").
			Comment("status of the channel configuration").
			GoType(enums.NotificationChannelStatus("")).
			Default(enums.NotificationChannelStatusEnabled.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.String("provider").
			Comment("provider service for the channel, e.g. sendgrid, mailgun for email or workspace name for slack").
			Optional(),
		field.String("destination").
			Comment("destination address or endpoint for the channel").
			Optional(),
		field.JSON("config", map[string]any{}).
			Comment("channel configuration payload").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Bool("enabled").
			Comment("whether this preference is enabled").
			Default(true).
			Annotations(
				entgql.OrderField("ENABLED"),
			),
		field.Enum("cadence").
			Comment("delivery cadence for this preference").
			GoType(enums.NotificationCadence("")).
			Default(enums.NotificationCadenceImmediate.String()),
		field.Enum("priority").
			Comment("optional priority override for this preference").
			GoType(enums.Priority("")).
			Optional(),
		field.Strings("topic_patterns").
			Comment("soiree topic names or wildcard patterns this preference applies to; empty means all").
			Optional(),
		field.JSON("topic_overrides", map[string]any{}).
			Comment("optional per-topic overrides (e.g. template_id, cadence, priority) keyed by soiree topic name").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("template_id").
			Comment("optional template to use by default for this preference (external channels only)").
			Optional(),
		field.Time("mute_until").
			Comment("mute notifications until this time").
			Optional().
			Nillable(),
		field.String("quiet_hours_start").
			Comment("start of quiet hours in HH:MM").
			Optional(),
		field.String("quiet_hours_end").
			Comment("end of quiet hours in HH:MM").
			Optional(),
		field.String("timezone").
			Comment("timezone for quiet hours and digests").
			Optional().
			Validate(func(s string) error {
				if s == "" {
					return nil
				}

				_, err := time.LoadLocation(s)

				return err
			}),
		field.Bool("is_default").
			Comment("whether this is the default config for the channel").
			Default(false),
		field.Time("verified_at").
			Comment("when the channel config was verified").
			Optional().
			Nillable(),
		field.Time("last_used_at").
			Comment("last time the channel config was used").
			Optional().
			Nillable(),
		field.Text("last_error").
			Comment("last error encountered while using the channel").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("additional preference metadata").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
	}
}

// Edges of the NotificationPreference.
func (n NotificationPreference) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: n,
			edgeSchema: User{},
			field:      "user_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: n,
			edgeSchema: NotificationTemplate{},
			field:      "template_id",
		}),
	}
}

// Indexes of the NotificationPreference.
func (NotificationPreference) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "user_id", "channel").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the NotificationPreference.
func (n NotificationPreference) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(n),
		},
	}.getMixins(n)
}

// Modules of the NotificationPreference.
func (NotificationPreference) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Policy of the NotificationPreference.
func (NotificationPreference) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfSelf(),
		),
		policy.WithMutationRules(
			rule.AllowIfSelf(),
		),
	)
}
