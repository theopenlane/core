package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
)

// TrustCenterSetting holds the schema definition for the TrustCenterSetting entity
type TrustCenterSetting struct {
	SchemaFuncs

	ent.Schema
}

const SchemaTrustCenterSetting = "trust_center_setting"

func (TrustCenterSetting) Name() string {
	return SchemaTrustCenterSetting
}
func (TrustCenterSetting) GetType() any {
	return TrustCenterSetting.Type
}

func (TrustCenterSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterSetting)
}

// Fields of the TrustCenterSetting
func (TrustCenterSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("trust_center_id").
			Comment("the ID of the trust center the settings belong to").
			NotEmpty(). // validate its not empty
			Optional(),
		field.String("title").
			Comment("title of the trust center").
			MaxLen(trustCenterNameMaxLen).
			Optional(),
		field.Text("overview").
			Comment("overview of the trust center").
			MaxLen(trustCenterDescriptionMaxLen).
			Optional(),
		field.String("logo_remote_url").
			Comment("URL of the logo").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("logo_local_file_id").
			Comment("The local logo file id, takes precedence over the logo remote URL").
			Optional().
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
		field.String("favicon_remote_url").
			Comment("URL of the favicon").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("favicon_local_file_id").
			Comment("The local favicon file id, takes precedence over the favicon remote URL").
			Optional().
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
		// Color/font settings
		field.Enum("theme_mode").
			Comment("Theme mode for the trust center").
			GoType(enums.TrustCenterThemeMode("")).
			Default(enums.TrustCenterThemeModeEasy.String()).
			Optional(),
		// Easy options
		field.String("primary_color").
			Comment("primary color for the trust center").
			Validate(validator.HexColorValidator).
			Optional(),
		// Advanced options
		field.String("font").
			Comment("font for the trust center").
			Optional(),
		field.String("foreground_color").
			Comment("foreground color for the trust center").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("background_color").
			Comment("background color for the trust center").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("accent_color").
			Comment("accent/brand color for the trust center").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("secondary_background_color").
			Comment("secondary background color for the trust center").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("secondary_foreground_color").
			Comment("secondary foreground color for the trust center").
			Validate(validator.HexColorValidator).
			Optional(),
		field.Enum("environment").
			GoType(enums.TrustCenterEnvironment("")).
			Default(enums.TrustCenterEnvironmentLive.String()).
			Immutable().
			Optional().
			Comment("environment of the trust center"),
		field.Bool("remove_branding").
			Comment("whether to remove branding from the trust center").
			Default(false).
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.String("company_domain").
			Comment("URL to the company's homepage").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("security_contact").
			Comment("email address for security contact").
			Optional().
			Nillable().
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.Bool("nda_approval_required").
			Comment("whether NDA requests require approval before being processed").
			Default(false).
			Optional(),
	}
}

// Mixin of the TrustCenterSetting
func (t TrustCenterSetting) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterSetting](t,
				withParents(TrustCenter{}),
			),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterSetting
func (t TrustCenterSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "logo_file",
			t:          File.Type,
			field:      "logo_local_file_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(File{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "favicon_file",
			t:          File.Type,
			field:      "favicon_local_file_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(File{}.Name()),
			},
		}),
	}
}

// Interceptors of the TrustCenterSetting
func (t TrustCenterSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}

// Hooks of the TrustCenterSetting
func (TrustCenterSetting) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterSetting(),
		hooks.HookTrustCenterSettingCreatePreview(),
	}
}

// Policy of the TrustCenterSetting
func (t TrustCenterSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfTrustCenterEditor(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterSettingMutation](),
		),
	)
}

func (TrustCenterSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trust_center_id", "environment").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (TrustCenterSetting) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

func (t TrustCenterSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}
