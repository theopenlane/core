package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
)

// ImpersonationEvent holds the schema definition for the ImpersonationEvent entity
type ImpersonationEvent struct {
	SchemaFuncs
	ent.Schema
}

// SchemaImpersonationEventis the name of the schema in snake case
const SchemaImpersonationEvent = "impersonation_event"

// Name is the name of the schema in snake case
func (ImpersonationEvent) Name() string {
	return SchemaImpersonationEvent
}

// GetType returns the type of the schema
func (ImpersonationEvent) GetType() any {
	return ImpersonationEvent.Type
}

// PluralName returns the plural name of the schema
func (ImpersonationEvent) PluralName() string {
	return pluralize.NewClient().Plural(SchemaImpersonationEvent)
}

// Fields of the ImpersonationEvent
func (ImpersonationEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("impersonation_type").
			Comment("Type of impersonation: SUPPORT, ADMIN, JOB").
			GoType(enums.ImpersonationType("")),
		field.Enum("action").
			Comment("Action for the impersonation event").
			GoType(enums.ImpersonationAction("")),
		field.String("reason").
			Optional().
			Comment("Reason for impersonation"),
		field.String("ip_address").
			Optional().
			Comment("IP address of the impersonator"),
		field.String("user_agent").
			Optional().
			Comment("User-Agent of the impersonator"),
		field.Strings("scopes").
			Optional().
			Comment("Granted scopes during impersonation"),
		field.String("user_id").
			Comment("Impersonator user id"),
		field.String("organization_id").
			Comment("id of the organization that is being impersonated"),
		field.String("target_user_id").
			Comment("id of the user that is being impersonated"),
	}
}

// Mixin of the ImpersonationEvent
func (i ImpersonationEvent) Mixin() []ent.Mixin {
	// graphql annotations are not required
	return mixinConfig{excludeAnnotations: true}.getMixins(i)
}

// Edges of the ImpersonationEvent
func (i ImpersonationEvent) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: User{},
			field:      "user_id",
			required:   true,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			name:       "target_user",
			t:          User.Type,
			field:      "target_user_id",
			required:   true,
			ref:        "targeted_impersonations",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: Organization{},
			field:      "organization_id",
			required:   true,
		}),
	}
}

// Indexes of the ImpersonationEvent
func (ImpersonationEvent) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ImpersonationEvent
func (ImpersonationEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the ImpersonationEvent
func (ImpersonationEvent) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the ImpersonationEvent
func (ImpersonationEvent) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Modules this schema has access to
func (ImpersonationEvent) Modules() []models.OrgModule {
	return []models.OrgModule{}
}
