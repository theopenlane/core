package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

// CustomDomain holds the schema definition for the CustomDomain
type CustomDomain struct {
	SchemaFuncs

	ent.Schema
}

const (
	// SchemaCustomDomain is the name of the CustomDomain schema.
	SchemaCustomDomain = "custom_domain"
	maxDomainNameLen   = 255
	maxTXTValueLen     = 64
	cloudflareIDMaxLen = 64
	maxStatusReasonLen = 255
)

// Name returns the name of the CustomDomain schema.
func (CustomDomain) Name() string {
	return SchemaCustomDomain
}

// GetType returns the type of the CustomDomain schema.
func (CustomDomain) GetType() any {
	return CustomDomain.Type
}

// PluralName returns the plural name of the CustomDomain schema.
func (CustomDomain) PluralName() string {
	return pluralize.NewClient().Plural(SchemaCustomDomain)
}

// Fields of the CustomDomain
func (CustomDomain) Fields() []ent.Field {
	return []ent.Field{
		field.String("cname_record").
			Comment("the name of the custom domain").
			Validate(validator.ValidateURL()).
			MaxLen(maxDomainNameLen).
			NotEmpty().
			Immutable().
			Annotations(
				entgql.OrderField("cname_record"),
			),
		field.String("mappable_domain_id").
			Comment("The mappable domain id that this custom domain maps to").
			NotEmpty().
			Immutable(),
		// TODO skip
		field.String("cloudflare_hostname_id").
			Comment("The ID of the custom domain in cloudflare").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Optional().
			MaxLen(cloudflareIDMaxLen),
		field.Enum("dns_verification_status").
			Comment("Status of the custom domain verification").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Default(enums.CustomDomainStatusPending.String()).
			GoType(enums.CustomDomainStatus("")),
		field.String("dns_verification_status_reason").
			Comment("Reason of the dns verification status, for giving the user diagnostic info").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			MaxLen(maxStatusReasonLen).
			Optional(),
		field.Enum("ssl_cert_status").
			Comment("Status of the ssl cert issuance").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Default(enums.CustomDomainStatusPending.String()).
			GoType(enums.CustomDomainStatus("")),
		field.String("ssl_cert_status_reason").
			Comment("Reason of the cert status, for giving the user diagnostic info").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			MaxLen(maxStatusReasonLen).
			Optional(),
		field.String("txt_record_subdomain").
			Comment("String to be prepended to the cname_record, used to evaluate domain ownership.").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Default("_olverify").
			Immutable().
			Optional().
			Deprecated("Field no longer used").
			MaxLen(16), //nolint:mnd
		field.String("txt_record_value").
			Comment("Hashed expected value of the TXT record. This is a random string that is generated on creation and is used to verify ownership of the domain.").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Immutable().
			DefaultFunc(func() string {
				str, err := hooks.GenerateDomainValidationSecret()
				if err != nil {
					// validators will catch this
					return ""
				}
				// This will get hashed by the hook before being stored in the DB
				// We want hash this in the hook so that we can send the unhashed
				// value back with the Create mutation response
				return str
			}).
			Optional().
			Deprecated("Field no longer used").
			MaxLen(maxTXTValueLen),
		field.String("status").
			Comment("Status of the custom domain verification").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional().
			Deprecated("Field no longer used").
			Default(enums.CustomDomainStatusPending.String()),
	}
}

// Mixin of the CustomDomain
func (e CustomDomain) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e, withSkipForSystemAdmin(true)),
		},
	}.getMixins()
}

// Edges of the CustomDomain
func (e CustomDomain) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: MappableDomain{},
			field:      "mappable_domain_id",
			required:   true,
			immutable:  true,
		}),
	}
}

// Indexes of the CustomDomain
func (CustomDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cname_record").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("owner_id").
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Policy of the CustomDomain
func (CustomDomain) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowQueryIfSystemAdmin(),
			policy.CheckOrgReadAccess(),
		),
		policy.WithOnMutationRules(
			ent.OpUpdateOne|ent.OpUpdate,
			rule.AllowMutationIfSystemAdmin(),
			privacy.AlwaysDenyRule(),
		),
		policy.WithOnMutationRules(
			ent.OpCreate|ent.OpDeleteOne|ent.OpDelete,
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (CustomDomain) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCustomDomain(),
	}
}
