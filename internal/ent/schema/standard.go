package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

// Standard defines the standard schema.
type Standard struct {
	SchemaFuncs

	ent.Schema
}

const SchemaStandard = "standard"

func (Standard) Name() string {
	return SchemaStandard
}

func (Standard) GetType() any {
	return Standard.Type
}

func (Standard) PluralName() string {
	return pluralize.NewClient().Plural(SchemaStandard)
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
		field.Text("description").
			Optional().
			Comment("long description of the standard with details of what is covered"),
		field.String("governing_body_logo_url").
			Comment("URL to the logo of the governing body").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
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
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
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
			Comment("indicates if the standard should be made available to all users, only for system owned standards"),
		field.Bool("free_to_use").
			Optional().
			Default(false).
			Comment("indicates if the standard is freely distributable under a trial license, only for system owned standards"),
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
func (s Standard) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Control{},
			annotations: []schema.Annotation{
				// skip the ability to create and update controls via the standard
				// TODO: (sfunk) implement permissions on parent edge to allow children to be created
				// and have the permissions added to fga
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			},
		}),
	}
}

// Mixin of the Standard
func (s Standard) Mixin() []ent.Mixin {
	return mixinConfig{
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s,
				withSkipForSystemAdmin(true), // allow empty owner_id for system admin
			),
			mixin.SystemOwnedMixin{},
		},
	}.getMixins()
}

// Hooks of the Standard
func (Standard) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookStandardPublicAccessTuples(),
		hooks.HookStandardCreate(),
	}
}

// Interceptors of the Standard
func (Standard) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.TraverseStandard(),
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
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
