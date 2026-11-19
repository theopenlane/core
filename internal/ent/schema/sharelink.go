package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/history"
)

const (
	// unlimitedShareLinkUses is the max uses value that places no cap on downloads
	unlimitedShareLinkUses = 0
	// defaultShareLinkMaxUses is the number of downloads a share link allows when none is requested
	defaultShareLinkMaxUses = unlimitedShareLinkUses
	// defaultShareLinkDuration is the lifetime of a share link created without an expiry
	defaultShareLinkDuration = 7 * 24 * time.Hour
)

// ShareLink holds the schema definition for the ShareLink entity, an external link granting time-limited
// downloads of its files
type ShareLink struct {
	SchemaFuncs

	ent.Schema
}

// SchemaShareLink is the name of the ShareLink schema
const SchemaShareLink = "share_link"

// Name returns the name of the ShareLink schema
func (ShareLink) Name() string {
	return SchemaShareLink
}

// GetType returns the type of the ShareLink schema
func (ShareLink) GetType() any {
	return ShareLink.Type
}

// PluralName returns the plural name of the ShareLink schema
func (ShareLink) PluralName() string {
	return pluralize.NewClient().Plural(SchemaShareLink)
}

// defaultShareLinkExpiry returns the expiry applied to a share link created without one
func defaultShareLinkExpiry() time.Time {
	return time.Now().Add(defaultShareLinkDuration)
}

// Fields of the ShareLink
func (ShareLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("a display name for the share link").
			Optional().
			Annotations(
				entgql.OrderField("name"),
			),
		field.Time("expires_at").
			Comment("when the share link stops serving downloads; defaults to seven days out and may be at most thirty days out").
			Default(defaultShareLinkExpiry).
			Immutable().
			Optional().
			Annotations(
				entgql.OrderField("expires_at"),
			),
		field.Int("max_uses").
			Comment("the number of downloads the share link allows, zero for unlimited").
			Default(defaultShareLinkMaxUses).
			Min(unlimitedShareLinkUses).
			Optional(),
		field.Int("use_count").
			Comment("the number of downloads served through the share link").
			Default(0).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			),
		field.Strings("recipient_emails").
			Comment("the email addresses the share link is bound to; when set a recipient must verify one of them before downloading").
			Immutable().
			Optional(),
		field.Bool("landing_page").
			Comment("whether the share URL opens a page listing the files instead of serving the file directly; always set for links with recipients or more than one file").
			Default(false).
			Immutable().
			Optional(),
		field.Bool("require_session").
			Comment("whether a download must present the credential of the caller that created the share link, who must still be able to read the file").
			Default(false).
			Immutable().
			Optional(),
		field.Bytes("code_hash").
			Comment("sha256 of the verification code currently issued to the recipient").
			Sensitive().
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
		field.Time("code_expires_at").
			Comment("when the current verification code stops being accepted").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
		field.Int("code_attempts").
			Comment("the number of failed attempts against the current verification code").
			Default(0).
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
		field.Time("revoked_at").
			Comment("when the share link was revoked").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			),
		field.String("token").
			Comment("the opaque share token carried by the share URL").
			Immutable().
			Sensitive().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
		field.Bytes("secret").
			Comment("the comparison secret to verify the share token's signature").
			Immutable().
			Sensitive().
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			),
	}
}

// Mixin of the ShareLink
func (s ShareLink) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
		},
	}.getMixins(s)
}

// Edges of the ShareLink
func (s ShareLink) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: File{},
			required:   true,
			immutable:  true,
			comment:    "the files the share link grants downloads of",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ShareLinkAccessEvent{},
			comment:    "the downloads served through the share link",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			},
		}),
	}
}

// Hooks of the ShareLink
func (ShareLink) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookShareLinkCreate(),
		hooks.HookShareLinkUpdate(),
	}
}

// Interceptors of the ShareLink
func (ShareLink) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorShareLinkURL(),
	}
}

// Policy of the ShareLink
func (ShareLink) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			// a share link grants downloads, so only members who can manage share links may read them
			policy.CheckOrgRelationAccess(rule.CanManageShareLinkRelation),
		),
		policy.WithMutationRules(
			policy.CheckOrgRelationAccess(rule.CanManageShareLinkRelation),
		),
	)
}

// Annotations of the ShareLink; history is excluded so the share token and code hash are never copied into a
// history table
func (ShareLink) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Modules of the ShareLink
func (ShareLink) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogAnyModule,
	}
}
