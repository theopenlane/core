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
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/entfga"
)

const (
	// Default watermark configuration values
	defaultWatermarkFontSize = 48.0
	defaultWatermarkOpacity  = 0.3
	defaultWatermarkRotation = 45.0
	maxWatermarkRotation     = 360.0
	watermarkTextMaxLen      = 255
	// grey
	defaultWatermarkColor = "#808080"
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
			MaxLen(watermarkTextMaxLen).
			Optional(),
		field.Float("font_size").
			Comment("font size of the watermark text").
			Default(defaultWatermarkFontSize).
			Optional(),
		field.Float("opacity").
			Comment("opacity of the watermark text").
			Default(defaultWatermarkOpacity).
			Min(0.0).
			Max(1.0).
			Optional(),
		field.Float("rotation").
			Comment("rotation of the watermark text").
			Default(defaultWatermarkRotation).
			Min(-maxWatermarkRotation).
			Max(maxWatermarkRotation).
			Optional(),
		field.String("color").
			Comment("color of the watermark text").
			Validate(validator.HexColorValidator).
			Default(defaultWatermarkColor).
			Optional(),
		field.Enum("font").
			Comment("font of the watermark text").
			GoType(enums.Font("")).
			Default(enums.FontHelvetica.String()).
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
				withOrganizationOwner(true),
			),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterWatermarkConfig
func (t TrustCenterWatermarkConfig) Edges() []ent.Edge {
	return []ent.Edge{
		nonUniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
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
	return []ent.Hook{
		hooks.HookTrustCenterWatermarkConfig(),
	}
}

// Policy of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
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
		entsql.Annotation{
			Checks: map[string]string{
				"text_or_logo_id_not_null": "(text IS NOT NULL) OR (logo_id IS NOT NULL)",
			},
		},
	}
}

// Interceptors of the TrustCenterWatermarkConfig
func (TrustCenterWatermarkConfig) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
