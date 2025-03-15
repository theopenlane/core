package schema

import (
	"net/url"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
)

// Standard defines the standard schema.
type Standard struct {
	ent.Schema
}

// Fields returns standard fields.
func (Standard) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			Comment("the long name of the standard body").
			Annotations(entx.FieldSearchable()),
		field.String("short_name").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("short_name"),
			).
			Comment("short name of the standard, e.g. SOC 2, ISO 27001, etc."),
		field.Text("framework").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("framework"),
			).
			Comment("unique identifier of the standard with version"),
		field.String("description").
			Optional().
			Comment("description of the standard"),
		field.String("governing_body_logo_url").
			Comment("URL to the logo of the governing body").
			MaxLen(urlMaxLen).
			Validate(func(s string) error {
				_, err := url.Parse(s)
				return err
			}).
			Optional(),
		field.String("governing_body").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("governing_body"),
			).
			Comment("governing body of the standard, e.g. AICPA, etc."),
		field.Strings("domains").
			Annotations(
				entx.FieldSearchable(),
			).
			Optional().
			Comment("domains the standard covers, e.g. availability, confidentiality, etc."),
		field.String("link").
			Optional().
			Comment("link to the official standard documentation"),
		field.Enum("status").
			GoType(enums.StandardStatus("")).
			Default(enums.StandardActive.String()).
			Optional().
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Comment("status of the standard - active, draft, and archived"),
		field.Bool("is_public").
			Optional().
			Default(false).
			Comment("indicates if the standard should be made available to all users, only for public standards"),
		field.Bool("free_to_use").
			Optional().
			Default(false).
			Comment("indicates if the standard is freely distributable under a trial license, only for public standards"),
		field.Bool("system_owned").
			Optional().
			Default(false).
			Comment("indicates if the standard is owned by the the openlane system"),
		field.String("standard_type").
			Annotations(
				entgql.OrderField("standard_type"),
			).
			Optional().
			Comment("type of the standard - cybersecurity, healthcare , financial, etc."),
		field.String("version").
			Optional().
			Comment("version of the standard"),
	}
}

// Edges of the Standard
func (Standard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controls", Control.Type).Annotations(entgql.RelayConnection()),
	}
}

// Mixin of the Standard
func (Standard) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		mixin.RevisionMixin{},
		NewOrgOwnedMixin(ObjectOwnedMixin{
			Ref:                      "standards",
			AllowEmptyForSystemAdmin: true, // allow empty org_id
		}),
	}
}

// Hooks of the Standard
func (Standard) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookStandardPublicAccessTuples(),
	}
}

// Interceptors of the Standard
func (Standard) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.TraverseStandard(),
	}
}

// Annotations of the Standard
func (Standard) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.MultiOrder(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}

// Policy of the Standard
func (Standard) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), // access is filtered by the traversal interceptor
		),
		policy.WithMutationRules(
			rule.SystemOwnedStandards(), // checks for the system owned field
			privacy.AlwaysDenyRule(),    // deny all other mutations for now
			// policy.CheckCreateAccess(), // TODO (sfunk): fix create access for org owned standards
			// privacy.AlwaysAllowRule(),
		),
	)
}
