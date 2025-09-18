package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
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
		field.Enum("visibility").
			GoType(enums.TrustCenterDocumentVisibility("")).
			Default(enums.TrustCenterDocumentVisibilityNotVisible.String()).
			Optional().
			Comment("visibility of the document"),
	}
}

// Mixin of the TrustCenterDoc
func (t TrustCenterDoc) Mixin() []ent.Mixin {
	return mixinConfig{}.getMixins(t)
}

// Edges of the TrustCenterDoc
func (t TrustCenterDoc) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: File{},
			field:      "file_id",
			comment:    "the file containing the document content",
		}),
	}
}

// Hooks of the TrustCenterDoc
func (TrustCenterDoc) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterDoc(),
		hooks.HookTrustCenterDocAuthz(),
	}
}

// Policy of the TrustCenterDoc
func (TrustCenterDoc) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
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
