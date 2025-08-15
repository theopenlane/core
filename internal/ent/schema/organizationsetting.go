package schema

import (
	"net/mail"
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// OrganizationSetting holds the schema definition for the OrganizationSetting entity
type OrganizationSetting struct {
	SchemaFuncs

	ent.Schema
}

// SchemaOrganizationSetting is the name of the OrganizationSetting schema.
const SchemaOrganizationSetting = "organization_setting"

// Name returns the name of the OrganizationSetting schema.
func (OrganizationSetting) Name() string {
	return SchemaOrganizationSetting
}

// GetType returns the type of the OrganizationSetting schema.
func (OrganizationSetting) GetType() any {
	return OrganizationSetting.Type
}

// PluralName returns the plural name of the OrganizationSetting schema.
func (OrganizationSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrganizationSetting)
}

// Fields of the OrganizationSetting
func (OrganizationSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Strings("domains").
			Comment("domains associated with the organization").
			Validate(validator.ValidateDomains()).
			Optional(),
		field.String("billing_contact").
			Comment("Name of the person to contact for billing").
			Optional(),
		field.String("billing_email").
			Comment("Email address of the person to contact for billing").
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}).
			Optional(),
		field.String("billing_phone").
			Comment("Phone number to contact for billing").
			Validate(func(phone string) error {
				regex := `^\+[1-9]{1}[0-9]{3,14}$`
				_, err := regexp.MatchString(regex, phone)
				return err
			}).
			Optional(),
		field.JSON("billing_address", models.Address{}).
			Comment("the billing address to send billing information to").
			Optional(),
		field.String("tax_identifier").
			Comment("Usually government-issued tax ID or business ID such as ABN in Australia").
			Optional(),
		field.Enum("geo_location").
			GoType(enums.Region("")).
			Comment("geographical location of the organization").
			Default(enums.Amer.String()).
			Optional(),
		field.String("organization_id").
			Comment("the ID of the organization the settings belong to").
			Optional(),
		field.Bool("billing_notifications_enabled").
			Comment("should we send email notifications related to billing").
			Default(true),
		field.Strings("allowed_email_domains").
			Comment("domains allowed to access the organization, if empty all domains are allowed").
			Validate(validator.ValidateDomains()).
			Optional(),
		field.Enum("identity_provider").
			Comment("SSO provider type for the organization").
			GoType(enums.SSOProvider("")).
			Optional().
			Default(string(enums.SSOProviderNone)),
		field.String("identity_provider_client_id").
			Comment("client ID for SSO integration").
			Nillable().
			Optional(),
		field.String("identity_provider_client_secret").
			Comment("client secret for SSO integration").
			Nillable().
			Optional(),
		field.String("identity_provider_metadata_endpoint").
			Comment("metadata URL for the SSO provider").
			Optional(),
		field.String("identity_provider_entity_id").
			Comment("SAML entity ID for the SSO provider").
			Optional(),
		field.String("oidc_discovery_endpoint").
			Comment("OIDC discovery URL for the SSO provider").
			Optional(),
		field.Bool("identity_provider_login_enforced").
			Comment("enforce SSO authentication for organization members").
			Default(false),
		field.String("compliance_webhook_token").
			Comment("unique token used to receive compliance webhook events").
			Unique().
			Optional().
			DefaultFunc(func() string {
				token := keygen.PrefixedSecret("tola_wsec")
				return token
			}),
	}
}

// Edges of the OrganizationSetting
func (o OrganizationSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: o,
			edgeSchema: Organization{},
			field:      "organization_id",
			ref:        "setting",
		}),
		defaultEdgeToWithPagination(o, File{}),
	}
}

// Annotations of the OrganizationSetting
func (o OrganizationSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("organization"),
	}
}

// Hooks of the OrganizationSetting
func (OrganizationSetting) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookOrganizationCreatePolicy(),
		hooks.HookOrganizationUpdatePolicy(),
	}
}

// Interceptors of the OrganizationSetting
func (o OrganizationSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorOrganizationSetting(),
	}
}

// Mixin of the OrganizationSetting
func (OrganizationSetting) Mixin() []ent.Mixin {
	return getDefaultMixins(OrganizationSetting{})
}

// Policy defines the privacy policy of the OrganizationSetting
func (OrganizationSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(), // access based on auth context
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.OrganizationSettingMutation](),
			policy.CheckOrgWriteAccess(), // access based on auth context
		),
	)
}

func (OrganizationSetting) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
