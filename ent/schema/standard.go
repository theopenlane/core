package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/interceptors"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/validator"
	"github.com/theopenlane/shared/directives"
	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/models"
)

// Standard defines the standard schema.
type Standard struct {
	SchemaFuncs

	ent.Schema
}

// SchemaStandard is the name of the standard schema.
const SchemaStandard = "standard"

// Name returns the name of the standard schema.
func (Standard) Name() string {
	return SchemaStandard
}

// GetType returns the type of the standard schema.
func (Standard) GetType() any {
	return Standard.Type
}

// PluralName returns the plural name of the standard schema.
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
			Annotations(
				directives.HiddenDirectiveAnnotation,
			).
			Comment("indicates if the standard should be made available to all users, only for system owned standards"),
		field.Bool("free_to_use").
			Optional().
			Default(false).
			Annotations(
				directives.HiddenDirectiveAnnotation,
			).
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
		field.String("logo_file_id").
			Comment("URL of the logo").
			Optional().
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
	}
}

// Edges of the Standard
func (s Standard) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Control{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: TrustCenterCompliance{},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: TrustCenterDoc{},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			name:       "logo_file",
			t:          File.Type,
			field:      "logo_file_id",
		}),
	}
}

// Mixin of the Standard
func (s Standard) Mixin() []ent.Mixin {
	return mixinConfig{
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s,
				withAllowAnonymousTrustCenterAccess(true),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(s)
}

// Hooks of the Standard
func (Standard) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookStandardPublicAccessTuples(),
		hooks.HookStandardCreate(),
		hooks.HookStandardFileUpload(),
		hooks.HookStandardDelete(),
		hook.On(
			hooks.OrgOwnedTuplesHook(),
			ent.OpCreate,
		),
	}
}

// Interceptors of the Standard
func (s Standard) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.TraverseStandard(),
	}
}

// Policy of the Standard
func (s Standard) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (Standard) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
		models.CatalogRiskManagementAddon,
		models.CatalogEntityManagementModule,
		models.CatalogTrustCenterModule,
	}
}
