package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

// EmailBranding holds the schema definition for the EmailBranding entity.
type EmailBranding struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEmailBranding is the name of the EmailBranding schema.
const SchemaEmailBranding = "email_branding"

// Name returns the name of the EmailBranding schema.
func (EmailBranding) Name() string {
	return SchemaEmailBranding
}

// GetType returns the type of the EmailBranding schema.
func (EmailBranding) GetType() any {
	return EmailBranding.Type
}

// PluralName returns the plural name of the EmailBranding schema.
func (EmailBranding) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEmailBranding)
}

// Fields of the EmailBranding.
func (EmailBranding) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("friendly name for this email branding configuration").
			NotEmpty().
			MaxLen(nameMaxLen).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("brand_name").
			Comment("brand name displayed in templates").
			MaxLen(trustCenterNameMaxLen).
			Optional(),
		field.String("logo_remote_url").
			Comment("URL of the brand logo for emails").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("primary_color").
			Comment("primary brand color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("secondary_color").
			Comment("secondary brand color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("background_color").
			Comment("background color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("text_color").
			Comment("text color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("button_color").
			Comment("button background color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("button_text_color").
			Comment("button text color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("link_color").
			Comment("link color for emails").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("font_family").
			Comment("font family for emails").
			Optional(),
		field.Bool("is_default").
			Comment("whether this is the default email branding for the organization").
			Default(false).
			Optional(),
	}
}

// Mixin of the EmailBranding.
func (e EmailBranding) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.EmailBranding](e,
				withParents(Organization{}),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
		},
	}.getMixins(e)
}

// Edges of the EmailBranding.
func (e EmailBranding) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Campaign{}),
		defaultEdgeToWithPagination(e, EmailTemplate{}),
	}
}

// Modules this schema has access to.
func (EmailBranding) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the EmailBranding.
func (EmailBranding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("organization"),
	}
}

// Policy of the EmailBranding.
func (EmailBranding) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
