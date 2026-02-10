package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
)

// Entity holds the schema definition for the Entity entity
type Entity struct {
	SchemaFuncs

	ent.Schema
}

// SchemaEntity is the name of the Entity schema.
const SchemaEntity = "entity"

// Name returns the name of the Entity schema.
func (Entity) Name() string {
	return SchemaEntity
}

// GetType returns the type of the Entity schema.
func (Entity) GetType() any {
	return Entity.Type
}

// PluralName returns the plural name of the Entity schema.
func (Entity) PluralName() string {
	return pluralize.NewClient().Plural(SchemaEntity)
}

// Fields of the Entity
func (Entity) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the entity").
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "citext",
			}).
			MinLen(minNameLength).
			Validate(validator.SpecialCharValidator).
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("display_name").
			Comment("The entity's displayed 'friendly' name").
			MaxLen(nameMaxLen).
			Optional().
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("display_name"),
			),
		field.String("description").
			Comment("An optional description of the entity").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Strings("domains").
			Comment("domains associated with the entity").
			Validate(validator.ValidateDomains()).
			Optional(),
		field.String("entity_type_id").
			Comment("The type of the entity").
			Optional(),
		field.String("status").
			Comment("status of the entity").
			Default("active").
			Annotations(
				entgql.OrderField("status"),
			).
			Optional(),
		field.Bool("approved_for_use").
			Comment("whether the entity is approved for use").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("approved_for_use"),
			),
		field.Strings("linked_asset_ids").
			Comment("asset identifiers linked to the entity").
			Default([]string{}).
			Optional(),
		field.Bool("has_soc2").
			Comment("whether the entity has an active SOC 2 report").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("has_soc2"),
			),
		field.Time("soc2_period_end").
			Comment("SOC 2 reporting period end date").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("soc2_period_end"),
			),
		field.Time("contract_start_date").
			Comment("start date for the entity contract").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("contract_start_date"),
			),
		field.Time("contract_end_date").
			Comment("end date for the entity contract").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("contract_end_date"),
			),
		field.Bool("auto_renews").
			Comment("whether the contract auto-renews").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("auto_renews"),
			),
		field.Int("termination_notice_days").
			Comment("number of days required for termination notice").
			Optional().
			Annotations(
				entgql.OrderField("termination_notice_days"),
			),
		field.Float("annual_spend").
			Comment("annual spend associated with the entity").
			Optional().
			Annotations(
				entgql.OrderField("annual_spend"),
			),
		field.String("spend_currency").
			Comment("the currency of the annual spend").
			Default("USD").
			Optional().
			Annotations(
				entgql.OrderField("spend_currency"),
			),
		field.String("billing_model").
			Comment("billing model for the entity relationship").
			Optional().
			Annotations(
				entgql.OrderField("billing_model"),
			),
		field.String("renewal_risk").
			Comment("renewal risk rating for the entity").
			Optional().
			Annotations(
				entgql.OrderField("renewal_risk"),
			),
		field.Bool("sso_enforced").
			Comment("whether SSO is enforced for the entity").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("sso_enforced"),
			),
		field.Bool("mfa_supported").
			Comment("whether MFA is supported by the entity").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("mfa_supported"),
			),
		field.Bool("mfa_enforced").
			Comment("whether MFA is enforced by the entity").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("mfa_enforced"),
			),
		field.String("status_page_url").
			Comment("status page URL for the entity").
			Validate(func(u string) error {
				if u == "" {
					return nil
				}

				return validator.ValidateURL()(u)
			}).
			Optional().
			Annotations(
				entgql.OrderField("status_page_url"),
			),
		field.Strings("provided_services").
			Comment("services provided by the entity").
			Optional().
			Default([]string{}),
		field.Strings("links").
			Comment("external links associated with the entity").
			Validate(validator.ValidateURLs()).
			Optional().
			Default([]string{}),
		field.String("risk_rating").
			Comment("the risk rating label for the entity").
			Optional().
			Annotations(
				entgql.OrderField("risk_rating"),
			),
		field.Int("risk_score").
			Comment("the risk score for the entity").
			Optional().
			Annotations(
				entgql.OrderField("risk_score"),
			),
		field.String("tier").
			Comment("the tier classification for the entity").
			Optional().
			Annotations(
				entgql.OrderField("tier"),
			),
		field.Enum("review_frequency").
			Comment("the cadence for reviewing the entity").
			GoType(enums.Frequency("")).
			Default(enums.FrequencyYearly.String()).
			Optional().
			Annotations(
				entgql.OrderField("REVIEW_FREQUENCY"),
			),
		field.Time("next_review_at").
			Comment("when the entity is due for review").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("next_review_at"),
			),
		field.Time("contract_renewal_at").
			Comment("when the entity contract is up for renewal").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("contract_renewal_at"),
			),
		field.JSON("vendor_metadata", map[string]any{}).
			Comment("vendor metadata such as additional enrichment info, company size, public, etc.").
			Optional(),
	}
}

// Mixin of the Entity
func (e Entity) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Entity](e,
				withParents(Organization{}, TrustCenterEntity{}),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			newResponsibilityMixin(e, withInternalOwner(), withReviewedBy(), withLastReviewedAt(), withReviewedByOrderField()),
			mixin.NewSystemOwnedMixin(),
			newCustomEnumMixin(e, withEnumFieldName("relationship_state")),
			newCustomEnumMixin(e, withEnumFieldName("security_questionnaire_status")),
			newCustomEnumMixin(e, withEnumFieldName("source_type")),
			newCustomEnumMixin(e, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(e, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(e)
}

// Edges of the Entity
func (e Entity) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(e, Contact{}),
		defaultEdgeToWithPagination(e, DocumentData{}),
		defaultEdgeToWithPagination(e, Note{}),
		defaultEdgeToWithPagination(e, File{}),
		defaultEdgeToWithPagination(e, Asset{}),
		defaultEdgeToWithPagination(e, Scan{}),
		defaultEdgeToWithPagination(e, Campaign{}),
		defaultEdgeToWithPagination(e, AssessmentResponse{}),
		defaultEdgeToWithPagination(e, Integration{}),
		defaultEdgeToWithPagination(e, Subprocessor{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: e,
			name:       "auth_methods",
			t:          CustomTypeEnum.Type,
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: e,
			name:       "employer_identity_holders",
			t:          IdentityHolder.Type,
			ref:        "employer",
		}),
		defaultEdgeFromWithPagination(e, IdentityHolder{}),
		defaultEdgeFromWithPagination(e, Platform{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: e,
			name:       "out_of_scope_platforms",
			t:          Platform.Type,
			ref:        "out_of_scope_vendors",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: e,
			name:       "source_platforms",
			t:          Platform.Type,
			ref:        "source_entities",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Platform{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: e,
			edgeSchema: EntityType{},
			field:      "entity_type_id",
		}),
	}
}

// Indexes of the Entity
func (Entity) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("reviewed_by_user_id"),
	}
}

// Hooks of the Entity
func (Entity) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEntityCreate(),
	}
}

// Policy of the Entity
func (e Entity) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.EntityMutation](),
		),
	)
}

// Annotations of the Entity
func (e Entity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

func (Entity) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogEntityManagementModule,
		models.CatalogComplianceModule,
	}
}
