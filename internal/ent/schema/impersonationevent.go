package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/models"
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
		field.String("impersonation_type").
			Comment("Type of impersonation: support, admin, job, etc."),
		field.Enum("action").Values("START", "STOP").
			Comment("Action for the impersonation event"),
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
			Comment("Impersonator organization id"),
		field.String("target_user_id").
			Comment("Target user id"),
	}
}

// Mixin of the ImpersonationEvent
func (i ImpersonationEvent) Mixin() []ent.Mixin {
	return getDefaultMixins(ImpersonationEvent{})
}

// Edges of the ImpersonationEvent
func (i ImpersonationEvent) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: User{},
			name:       "impersonator",
			field:      "user_id",
		}),
		// Event targets a user (single target per event)
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: User{},
			name:       "target_user",
			field:      "target_user_id",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: i,
			edgeSchema: Organization{},
			name:       "organization",
			field:      "organization_id",
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

// Policy of the ImpersonationEvent
func (ImpersonationEvent) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithMutationRules(
			// add mutation rules here, the below is the recommended default
			policy.CheckCreateAccess(),
			// this needs to be commented out for the first run that had the entfga annotation
			// the first run will generate the functions required based on the entfa annotation
			// entfga.CheckEditAccess[*generated.ImpersonationEventMutation](),
		),
	)
}
