package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// DNSVerification holds the schema definition for the DNSVerification
type DNSVerification struct {
	SchemaFuncs

	ent.Schema
}

const (
	// SchemaDNSVerification is the name of the DNSVerification schema.
	SchemaDNSVerification = "dns_verification"
	maxTXTValueLen        = 64
	cloudflareIDMaxLen    = 64
	maxStatusReasonLen    = 255
)

// Name returns the name of the DNSVerification schema.
func (DNSVerification) Name() string {
	return SchemaDNSVerification
}

// GetType returns the type of the DNSVerification schema.
func (DNSVerification) GetType() any {
	return DNSVerification.Type
}

// PluralName returns the plural name of the DNSVerification schema.
func (DNSVerification) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDNSVerification)
}

// Fields of the DNSVerification
func (DNSVerification) Fields() []ent.Field {
	return []ent.Field{
		field.String("cloudflare_hostname_id").
			Comment("The ID of the custom domain in cloudflare").
			Immutable().
			NotEmpty().
			MaxLen(cloudflareIDMaxLen),
		field.String("dns_txt_record").
			Comment("the name of the dns txt record").
			MaxLen(maxDomainNameLen).
			NotEmpty(),
		field.String("dns_txt_value").
			Comment("the expected value of the dns txt record").
			MaxLen(maxTXTValueLen).
			NotEmpty(),
		field.Enum("dns_verification_status").
			Comment("Status of the domain verification").
			Default(string(enums.DNSVerificationStatusPending)).
			GoType(enums.DNSVerificationStatus("")),
		field.String("dns_verification_status_reason").
			Comment("Reason of the dns verification status, for giving the user diagnostic info").
			MaxLen(maxStatusReasonLen).
			Optional(),
		field.String("acme_challenge_path").
			Comment("Path under /.well-known/acme-challenge/ to serve the ACME challenge").
			MaxLen(maxDomainNameLen).
			Optional(),
		field.String("expected_acme_challenge_value").
			Comment("the expected value of the acme challenge record").
			MaxLen(maxTXTValueLen).
			Optional(),
		field.Enum("acme_challenge_status").
			Comment("Status of the ACME challenge validation").
			Default(string(enums.SSLVerificationStatusInitializing)).
			GoType(enums.SSLVerificationStatus("")),
		field.String("acme_challenge_status_reason").
			Comment("Reason of the ACME status, for giving the user diagnostic info").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			MaxLen(maxStatusReasonLen).
			Optional(),
	}
}

// Mixin of the DNSVerification
func (e DNSVerification) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e, withSkipForSystemAdmin(true)),
		},
	}.getMixins(e)
}

// Edges of the DNSVerification
func (e DNSVerification) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, CustomDomain{}),
	}
}

// Indexes of the DNSVerification
func (DNSVerification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cloudflare_hostname_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Hooks of the DNSVerification
func (DNSVerification) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookDNSVerificationDelete(),
	}
}

// Modules of the DNSVerification
func (DNSVerification) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Policy of the DNSVerification
func (DNSVerification) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
		),
	)
}
