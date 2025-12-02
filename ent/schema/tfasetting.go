package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/shared/models"
)

// TFASetting holds the schema definition for the TFASetting entity
type TFASetting struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTFASetting is the name of the TFASetting schema.
const SchemaTFASetting = "tfa_setting"

// Name returns the name of the TFASetting schema.
func (TFASetting) Name() string {
	return SchemaTFASetting
}

// GetType returns the type of the TFASetting schema.
func (TFASetting) GetType() any {
	return TFASetting.Type
}

// PluralName returns the plural name of the TFASetting schema.
func (TFASetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTFASetting)
}

// Fields of the TFASetting
func (TFASetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("tfa_secret").
			Comment("TFA secret for the user").
			Annotations(
				entgql.Skip(entgql.SkipAll),
			).
			Optional().
			Nillable(),
		field.Bool("verified").
			Comment("specifies if the TFA device has been verified").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Default(false),
		field.Strings("recovery_codes").
			Comment("recovery codes for 2fa").
			Annotations(
				entgql.Skip(entgql.SkipAll),
			).
			Optional(),
		field.Bool("phone_otp_allowed").
			Comment("specifies a user may complete authentication by verifying an OTP code delivered through SMS").
			Optional().
			Annotations(
				// skip until feature is implemented
				entgql.Skip(entgql.SkipAll),
			).
			Default(false),
		field.Bool("email_otp_allowed").
			Comment("specifies a user may complete authentication by verifying an OTP code delivered through email").
			Optional().
			Annotations(
				// skip until feature is implemented
				entgql.Skip(entgql.SkipAll),
			).
			Default(false),
		field.Bool("totp_allowed").
			Comment("specifies a user may complete authentication by verifying a TOTP code delivered through an authenticator app").
			Optional().
			Default(false),
	}
}

// Mixin of the TFASetting
func (t TFASetting) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newUserOwnedMixin(t,
				withOptionalUser(),
				withSoftDeleteIndex(),
			),
		},
	}.getMixins(t)
}

// Hooks of the TFASetting
func (TFASetting) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEnableTFA(), // sets 2fa secret if totp is allowed
		hooks.HookVerifyTFA(), // generates recovery codes and enables TFA on a user settings

	}
}

func (TFASetting) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the TFASetting
func (t TFASetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Policy of the TFASetting
func (t TFASetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfSelf(),
		),
		policy.WithMutationRules(
			rule.AllowIfSelf(),
		),
	)
}
