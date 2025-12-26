package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"

	"github.com/theopenlane/entx/accessmap"
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
		field.String("dns_verification_id").
			Comment("The ID of the dns verification record").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Optional(),
	}
}

// Mixin of the CustomDomain
func (e CustomDomain) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e, withAllowAnonymousTrustCenterAccess(true)),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(e)
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
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: DNSVerification{},
			field:      "dns_verification_id",
			required:   false,
			immutable:  false,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
	}
}

// Indexes of the CustomDomain
func (CustomDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cname_record").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Policy of the CustomDomain
func (e CustomDomain) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowQueryIfSystemAdmin(),
		),
		policy.WithOnMutationRules(
			ent.OpCreate|ent.OpDeleteOne|ent.OpDelete|ent.OpUpdateOne|ent.OpUpdate,
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Hooks of the CustomDomain
func (CustomDomain) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreateCustomDomain(),
		hooks.HookDeleteCustomDomain(),
	}
}

func (CustomDomain) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}
