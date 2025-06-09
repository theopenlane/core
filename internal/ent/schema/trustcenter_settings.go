package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/iam/entfga"
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
		field.String("logo_url").
			Comment("logo url for the trust center").
			MaxLen(trustCenterURLMaxLen).
			Validate(validator.ValidateURL()).
			Optional(),
		field.String("favicon_url").
			Comment("favicon url for the trust center").
			MaxLen(trustCenterURLMaxLen).
			Validate(validator.ValidateURL()).
			Optional(),
		field.String("primary_color").
			Comment("primary color for the trust center").
			Optional(),
	}
}

// Mixin of the TrustCenterSetting
func (t TrustCenterSetting) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
	}.getMixins()
}

// Edges of the TrustCenterSetting
func (t TrustCenterSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			ref:        "setting",
		}),
	}
}

// Interceptors of the TrustCenterSetting
func (TrustCenterSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterSetting(),
	}
}

// Hooks of the TrustCenterSetting
func (TrustCenterSetting) Hooks() []ent.Hook {
	return []ent.Hook{}
}

func (TrustCenterSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.TrustCenterSettingMutation](),
		),
	)
}

func (TrustCenterSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (TrustCenterSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}
