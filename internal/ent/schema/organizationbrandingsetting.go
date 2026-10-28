package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/ent/validator"
)

type OrganizationBrandingSetting struct {
	SchemaFuncs

	ent.Schema
}

const SchemaOrganizationBrandingSetting = "organization_branding_setting"

func (OrganizationBrandingSetting) Name() string {
	return SchemaOrganizationBrandingSetting
}

func (OrganizationBrandingSetting) GetType() any {
	return OrganizationBrandingSetting.Type
}

func (OrganizationBrandingSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaOrganizationBrandingSetting)
}

func (OrganizationBrandingSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("organization_id").
			Comment("the ID of the organization the settings belong to").
			Optional(),
		field.Bool("remove_assessment_branding").
			Comment("whether to hide the Openlane branding on the Organization's assessments").
			Default(false).
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.String("primary_color").
			Comment("the primary brand color of the Organization").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("secondary_color").
			Comment("the secondary brand color of the Organization").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("background_color").
			Comment("the background color of the Organization").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("text_color").
			Comment("the text color of the Organization").
			Validate(validator.HexColorValidator).
			Optional(),
		field.String("font_family").
			Comment("the font family of the Organization").
			Optional(),
	}
}

func (OrganizationBrandingSetting) Mixin() []ent.Mixin {
	return getDefaultMixins(OrganizationBrandingSetting{})
}

func (b OrganizationBrandingSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: b,
			edgeSchema: Organization{},
			field:      "organization_id",
			ref:        "branding_setting",
		}),
	}
}

func (OrganizationBrandingSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (OrganizationBrandingSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
			rule.AllowQueryIfSystemAdmin(),
		),
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (OrganizationBrandingSetting) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

func (OrganizationBrandingSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.FileCategory(SchemaOrganizationBrandingSetting),
		entfga.SettingsChecks(Organization{}.Name()),
		entx.FGACrudSkip(entx.SkipDelete | entx.SkipCreate),
		entx.FGACrudParent(Organization{}.Name()),
	}
}
