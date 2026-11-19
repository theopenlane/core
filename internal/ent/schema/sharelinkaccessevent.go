package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
)

// ShareLinkAccessEvent holds the schema definition for the ShareLinkAccessEvent entity, one download served
// through a share link
type ShareLinkAccessEvent struct {
	SchemaFuncs

	ent.Schema
}

// SchemaShareLinkAccessEvent is the name of the ShareLinkAccessEvent schema
const SchemaShareLinkAccessEvent = "share_link_access_event"

// Name returns the name of the ShareLinkAccessEvent schema
func (ShareLinkAccessEvent) Name() string {
	return SchemaShareLinkAccessEvent
}

// GetType returns the type of the ShareLinkAccessEvent schema
func (ShareLinkAccessEvent) GetType() any {
	return ShareLinkAccessEvent.Type
}

// PluralName returns the plural name of the ShareLinkAccessEvent schema
func (ShareLinkAccessEvent) PluralName() string {
	return pluralize.NewClient().Plural(SchemaShareLinkAccessEvent)
}

// Fields of the ShareLinkAccessEvent
func (ShareLinkAccessEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("share_link_id").
			Comment("the share link the download was served through").
			NotEmpty().
			Immutable(),
		field.String("file_id").
			Comment("the file that was downloaded").
			NotEmpty().
			Immutable(),
		field.String("ip_address").
			Comment("the client address the download was served to").
			Immutable().
			Optional(),
		field.String("user_agent").
			Comment("the client user agent the download was served to").
			Immutable().
			Optional(),
	}
}

// Mixin of the ShareLinkAccessEvent
func (s ShareLinkAccessEvent) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
		},
	}.getMixins(s)
}

// Edges of the ShareLinkAccessEvent
func (s ShareLinkAccessEvent) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ShareLink{},
			field:      "share_link_id",
			required:   true,
			immutable:  true,
			comment:    "the share link the download was served through",
		}),
	}
}

// Policy of the ShareLinkAccessEvent; events are written only by the share consume endpoint as an internal caller
func (ShareLinkAccessEvent) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			// access events carry client addresses, so only members who can manage share links may read them
			policy.CheckOrgRelationAccess(rule.CanManageShareLinkRelation),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
		),
	)
}

// Annotations of the ShareLinkAccessEvent; events are read through the share link and have no mutations
func (ShareLinkAccessEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		entx.FGACrudSkip(entx.SkipCreate | entx.SkipUpdate | entx.SkipDelete),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Modules of the ShareLinkAccessEvent
func (ShareLinkAccessEvent) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogAnyModule,
	}
}
