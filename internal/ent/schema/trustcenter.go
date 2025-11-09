package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
)

const (
	trustCenterNameMaxLen        = 160
	trustCenterDescriptionMaxLen = 1024
	trustCenterURLMaxLen         = 2048
)

// TrustCenter holds the schema definition for the TrustCenter entity
type TrustCenter struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenter is the name of the TrustCenter schema.
const SchemaTrustCenter = "trust_center"

// Name returns the name of the TrustCenter schema.
func (TrustCenter) Name() string {
	return SchemaTrustCenter
}

// GetType returns the type of the TrustCenter schema.
func (TrustCenter) GetType() any {
	return TrustCenter.Type
}

// PluralName returns the plural name of the TrustCenter schema.
func (TrustCenter) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenter)
}

// Fields of the TrustCenter
func (TrustCenter) Fields() []ent.Field {
	return []ent.Field{
		field.String("slug").
			Comment("Slug for the trust center").
			MaxLen(trustCenterNameMaxLen).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Optional(),
		field.String("custom_domain_id").
			Comment("custom domain id for the trust center").
			Optional(),
		field.String("pirsch_domain_id").
			Comment("Pirsch domain ID").
			Optional(),
		field.String("pirsch_identification_code").
			Comment("Pirsch ID code").
			Optional(),
	}
}

// Mixin of the TrustCenter
func (t TrustCenter) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t, withAllowAnonymousTrustCenterAccess(true), withSkipForSystemAdmin(true)),
		},
	}.getMixins(t)
}

// Edges of the TrustCenter
func (t TrustCenter) Edges() []ent.Edge {
	return []ent.Edge{
		// Add relationships to other entities as needed
		// Example: defaultEdgeToWithPagination(t, File{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: CustomDomain{},
			field:      "custom_domain_id",
			required:   false,
			immutable:  false,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "setting",
			t:          TrustCenterSetting.Type,
			annotations: []schema.Annotation{
				entx.CascadeAnnotationField("TrustCenter"),
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "watermark_config",
			t:          TrustCenterWatermarkConfig.Type,
			annotations: []schema.Annotation{
				entx.CascadeAnnotationField("TrustCenter"),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    t,
			name:          "subprocessors",
			edgeSchema:    TrustCenterSubprocessor{},
			cascadeDelete: "TrustCenter",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Organization{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    t,
			name:          "documents",
			edgeSchema:    TrustCenterDoc{},
			cascadeDelete: "TrustCenter",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    t,
			name:          "compliance",
			edgeSchema:    TrustCenterCompliance{},
			cascadeDelete: "TrustCenter",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    t,
			name:          "templates",
			edgeSchema:    Template{},
			cascadeDelete: "TrustCenter",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			name:       "posts",
			t:          Note.Type,
			comment:    "posts for the trust center feed",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
	}
}

// Hooks of the TrustCenter
func (TrustCenter) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenter(),
		hook.On(
			hooks.OrgOwnedTuplesHook(),
			ent.OpCreate,
		),
	}
}

// Policy of the TrustCenter
func (t TrustCenter) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowQueryIfSystemAdmin(),
		),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Indexes of the TrustCenter
func (TrustCenter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("slug").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (TrustCenter) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Annotations of the TrustCenter
func (t TrustCenter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the TrustCenter
func (t TrustCenter) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenter(),
	}
}
