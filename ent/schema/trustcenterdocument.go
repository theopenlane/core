package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/interceptors"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
)

// TrustCenterDoc holds the schema definition for the TrustCenterDoc entity
type TrustCenterDoc struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterDoc is the name of the TrustCenterDoc schema.
const SchemaTrustCenterDoc = "trust_center_doc"

// Name returns the name of the TrustCenterDoc schema.
func (TrustCenterDoc) Name() string {
	return SchemaTrustCenterDoc
}

// GetType returns the type of the TrustCenterDoc schema.
func (TrustCenterDoc) GetType() any {
	return TrustCenterDoc.Type
}

// PluralName returns the plural name of the TrustCenterDoc schema.
func (TrustCenterDoc) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterDoc)
}

// Fields of the TrustCenterDoc
func (TrustCenterDoc) Fields() []ent.Field {
	return []ent.Field{
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Optional(),
		field.String("title").
			Comment("title of the document").
			NotEmpty(),
		field.String("category").
			Comment("category of the document").
			NotEmpty(),
		field.String("file_id").
			Comment("ID of the file containing the document").
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional().
			Nillable(),
		field.String("original_file_id").
			Comment("ID of the file containing the document, before any watermarking").
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional().
			Nillable(),
		field.Bool("watermarking_enabled").
			Default(false).
			Comment("whether watermarking is enabled for the document. this will only take effect if watermarking is configured for the trust center"),
		field.Enum("watermark_status").
			GoType(enums.WatermarkStatus("")).
			Default(enums.WatermarkStatusDisabled.String()).
			Optional().
			Comment("status of the watermarking"),
		field.Enum("visibility").
			GoType(enums.TrustCenterDocumentVisibility("")).
			Default(enums.TrustCenterDocumentVisibilityNotVisible.String()).
			Optional().
			Comment("visibility of the document"),
		field.String("standard_id").
			Comment("ID of the standard").
			NotEmpty().
			Optional(),
	}
}

// Mixin of the TrustCenterDoc
func (t TrustCenterDoc) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterDoc](t,
				withParents(TrustCenter{}),
			),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterDoc
func (t TrustCenterDoc) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Standard{},
			field:      "standard_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Standard{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: File{},
			field:      "file_id",
			comment:    "the file containing the document content",
		}),
		uniqueEdgeTo(&edgeDefinition{
			name:       "original_file",
			fromSchema: t,
			t:          File.Type,
			field:      "original_file_id",
			comment:    "the file containing the document content, pre watermarking",
		}),
	}
}

// Hooks of the TrustCenterDoc
func (TrustCenterDoc) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookCreateTrustCenterDoc(),
		hooks.HookUpdateTrustCenterDoc(),
	}
}

// Policy of the TrustCenterDoc
func (TrustCenterDoc) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterDocMutation](),
		),
	)
}

func (TrustCenterDoc) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Indexes of the TrustCenterDoc
func (TrustCenterDoc) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterDoc
func (TrustCenterDoc) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the TrustCenterDoc
func (TrustCenterDoc) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
