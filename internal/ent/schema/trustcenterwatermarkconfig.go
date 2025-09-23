package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/entfga"
)

// TrustCenterWatermarkConfig holds the schema definition for the TrustCenterWatermarkConfig entity
type TrustCenterWatermarkConfig struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterWatermarkConfig is the name of the TrustCenterWatermarkConfig schema.
const SchemaTrustCenterWatermarkConfig = "trust_center_watermark_config"

// Name returns the name of the TrustCenterWatermarkConfig schema.
func (TrustCenterWatermarkConfig) Name() string {
	return SchemaTrustCenterWatermarkConfig
}

// GetType returns the type of the TrustCenterWatermarkConfig schema.
func (TrustCenterWatermarkConfig) GetType() any {
	return TrustCenterWatermarkConfig.Type
}

// PluralName returns the plural name of the TrustCenterWatermarkConfig schema.
func (TrustCenterWatermarkConfig) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterWatermarkConfig)
}

// Fields of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Optional(),
		field.String("logo_id").
			Comment("ID of the file containing the document").
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional().
			Nillable(),
		field.String("text").
			Comment("text to watermark the document with").
			Optional(),
		field.Float("font_size").
			Comment("font size of the watermark text").
			Default(48.0).
			Optional(),
		field.Float("opacity").
			Comment("opacity of the watermark text").
			Default(0.3).
			Min(0.0).
			Max(1.0).
			Optional(),
		field.Float("rotation").
			Comment("rotation of the watermark text").
			Default(45.0).
			Min(-360.0).
			Max(360.0).
			Optional(),
		field.String("color").
			Comment("color of the watermark text").
			Optional(),
		field.String("font").
			Comment("font of the watermark text").
			Optional(),
	}
}

// Mixin of the TrustCenterWatermarkConfig
func (t TrustCenterWatermarkConfig) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterWatermarkConfig](t,
				withParents(TrustCenter{}),
			),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterWatermarkConfig
func (t TrustCenterWatermarkConfig) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			ref:        "watermark_config",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: File{},
			field:      "logo_id",
			comment:    "the file containing the image for watermarking, if applicable",
		}),
	}
}

// Hooks of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.TrustCenterWatermarkConfigMutation](),
		),
	)
}

func (TrustCenterWatermarkConfig) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Indexes of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
