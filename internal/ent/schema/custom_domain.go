package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
)

// CustomDomain holds the schema definition for the CustomDomain
type CustomDomain struct {
	SchemaFuncs

	ent.Schema
}

const (
	SchemaCustomDomain = "custom_domain"
	maxDomainNameLen   = 255
	maxTXTValueLen     = 64
)

func (CustomDomain) Name() string {
	return SchemaCustomDomain
}

func (CustomDomain) GetType() any {
	return CustomDomain.Type
}

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
		field.String("txt_record_subdomain").
			Comment("String to be prepended to the cname_record, used to evaluate domain ownership.").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Default("_olverify").
			NotEmpty().
			Immutable().
			MaxLen(16),
		field.String("txt_record_value").
			Comment("Hashed expected value of the TXT record. This is a random string that is generated on creation and is used to verify ownership of the domain.").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			NotEmpty().
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
			MaxLen(maxTXTValueLen),
		field.Enum("status").
			Comment("Status of the custom domain verification").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			GoType(enums.CustomDomainStatus("")).
			Default(enums.CustomDomainStatusPending.String()),
	}
}

// Mixin of the CustomDomain
func (e CustomDomain) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(e),
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
			policy.CheckOrgReadAccess(),
		),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (CustomDomain) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCustomDomain(),
	}
}
