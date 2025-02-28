package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/entx"
	emixin "github.com/theopenlane/entx/mixin"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
			).
			Comment("the long name of the standard body").
			Annotations(entx.FieldSearchable()),
		field.String("short_name").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("short name of the standard, e.g. SOC 2, ISO 27001, etc."),
		field.Text("framework").
			Optional().
			Comment("unique identifier of the standard with version"),
		field.String("description").
			Optional().
			Comment("description of the standard"),
		field.String("governing_body").
			Optional().
			Comment("governing body of the standard, e.g. AICPA, etc."),
		field.Strings("domains").
			Optional().
			Comment("domains the standard covers, e.g. availability, confidentiality, etc."),
		field.String("link").
			Optional().
			Comment("link to the official standard documentation"),
		field.String("status").
			Optional().
			Comment("status of the standard - active, deprecated, etc."),
		field.Bool("is_public").
			Optional().
			Default(false).
			// don't expose these fields in the mutation input
			Annotations(entgql.Skip(^entgql.SkipType)).
			Comment("indicates if the standard should be made available to all users, only for public standards"),
		field.Bool("free_to_use").
			Optional().
			Default(false).
			// don't expose these fields in the mutation input
			Annotations(entgql.Skip(^entgql.SkipType)).
			Comment("indicates if the standard is freely distributable under a trial license, only for public standards"),
		field.Bool("system_owned").
			Optional().
			Default(false).
			// don't expose these fields in the mutation input
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput)).
			Comment("indicates if the standard is owned by the the openlane system"),
		field.String("standard_type").
			Optional().
			Comment("type of the standard - security, privacy, etc."),
		field.String("version").
			Optional().
			Comment("version of the standard"),
		field.String("revision").
			Optional().
			Comment("internal revision of the standard"),
	}
}

// Edges of the Standard
func (Standard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("controls", Control.Type),
	}
}

// Mixin of the Standard
func (Standard) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		NewOrgOwnMixinWithRef("standards"),
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
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}

// Policy of the Standard
func (Standard) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			// policy.CheckCreateAccess(),
			privacy.AlwaysAllowRule(),
		),
	)
}
