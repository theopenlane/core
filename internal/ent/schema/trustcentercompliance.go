package schema

// TrustCenterCompliance represents compliance with a framework, and associates
// it with the organization's trust center When implemented, this will have a
// pointer to a "program" object and its "standard" framework

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
)

// TrustCenterCompliance holds the schema definition for the TrustCenterCompliance entity
type TrustCenterCompliance struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterCompliance is the name of the TrustCenterCompliance schema.
const SchemaTrustCenterCompliance = "trust_center_compliance"

// Name returns the name of the TrustCenterCompliance schema.
func (TrustCenterCompliance) Name() string {
	return SchemaTrustCenterCompliance
}

// GetType returns the type of the TrustCenterCompliance schema.
func (TrustCenterCompliance) GetType() any {
	return TrustCenterCompliance.Type
}

// PluralName returns the plural name of the TrustCenterCompliance schema.
func (TrustCenterCompliance) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterCompliance)
}

// Fields of the TrustCenterCompliance
func (TrustCenterCompliance) Fields() []ent.Field {
	return []ent.Field{
		field.String("standard_id").
			Comment("ID of the standard").
			NotEmpty(),
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Immutable().
			Optional(),
	}
}

// Mixin of the TrustCenterCompliance
func (t TrustCenterCompliance) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterCompliance](t,
				withParents(TrustCenter{}),
				withAllowAnonymousTrustCenterAccess(true),
			),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterCompliance
func (t TrustCenterCompliance) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			immutable:  true,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Standard{},
			field:      "standard_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Standard{}.Name()),
			},
		}),
	}
}

// Modules this schema has access to
func (TrustCenterCompliance) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Hooks of the TrustCenterCompliance
func (TrustCenterCompliance) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterComplianceAuthz(),
	}
}

// Policy of the TrustCenterCompliance
func (TrustCenterCompliance) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithOnMutationRules(ent.OpCreate,
			policy.CheckCreateAccess(),
		),
		policy.WithMutationRules(
			rule.AllowIfTrustCenterEditor(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterComplianceMutation](),
		),
	)
}

// Indexes of the TrustCenterCompliance
func (TrustCenterCompliance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("standard_id", "trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the TrustCenterCompliance
func (TrustCenterCompliance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}

// Interceptors of the TrustCenterCompliance
func (TrustCenterCompliance) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
