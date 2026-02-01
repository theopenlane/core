package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// TrustCenterSubprocessor holds the schema definition for the TrustCenterSubprocessor entity
type TrustCenterSubprocessor struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterSubprocessor is the name of the TrustCenterSubprocessor schema.
const SchemaTrustCenterSubprocessor = "trust_center_subprocessor"

// Name returns the name of the TrustCenterSubprocessor schema.
func (TrustCenterSubprocessor) Name() string {
	return SchemaTrustCenterSubprocessor
}

// GetType returns the type of the TrustCenterSubprocessor schema.
func (TrustCenterSubprocessor) GetType() any {
	return TrustCenterSubprocessor.Type
}

// PluralName returns the plural name of the TrustCenterSubprocessor schema.
func (TrustCenterSubprocessor) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterSubprocessor)
}

// Fields of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Fields() []ent.Field {
	return []ent.Field{
		field.String("subprocessor_id").
			Comment("ID of the subprocessor").
			Immutable().
			NotEmpty(),
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Immutable().
			Optional(),
		field.JSON("countries", []string{}).
			Comment("country codes or country where the subprocessor is located").
			Optional(),
	}
}

// Mixin of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterSubprocessor](t,
				withParents(TrustCenter{}),
				withAllowAnonymousTrustCenterAccess(true),
			),
			newCustomEnumMixin(t),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			immutable:  true,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Subprocessor{},
			immutable:  true,
			field:      "subprocessor_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Organization{}.Name()),
			},
		}),
	}
}

// Hooks of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterSubprocessor(),
	}
}

// Policy of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithOnMutationRules(ent.OpCreate,
			policy.CheckCreateAccess(),
		),
		policy.WithMutationRules(
			rule.AllowIfTrustCenterEditor(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterSubprocessorMutation](),
		),
	)
}

// Indexes of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subprocessor_id", "trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (TrustCenterSubprocessor) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Annotations of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
		entx.NewExportable(),
	}
}

// Interceptors of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
