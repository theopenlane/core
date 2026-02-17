package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/iam/entfga"
)

// TrustCenterFAQ holds the schema definition for the TrustCenterFAQ entity
type TrustCenterFAQ struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterFAQ is the name of the TrustCenterFAQ schema.
const SchemaTrustCenterFAQ = "trust_center_faq"

// Name returns the name of the TrustCenterFAQ schema.
func (TrustCenterFAQ) Name() string {
	return SchemaTrustCenterFAQ
}

// GetType returns the type of the TrustCenterFAQ schema.
func (TrustCenterFAQ) GetType() any {
	return TrustCenterFAQ.Type
}

// PluralName returns the plural name of the TrustCenterFAQ schema.
func (TrustCenterFAQ) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterFAQ)
}

// Fields of the TrustCenterFAQ
func (TrustCenterFAQ) Fields() []ent.Field {
	return []ent.Field{
		field.String("reference_link").
			Comment("optional reference link for the FAQ").
			Validate(validator.ValidateURL()).
			Optional(),
		field.Int("display_order").
			Comment("display order of the FAQ").
			Default(0).
			Optional(),
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Immutable().
			Optional(),
	}
}

// Mixin of the TrustCenterFAQ
func (t TrustCenterFAQ) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterFAQ](t,
				withParents(TrustCenter{}),
				withAllowAnonymousTrustCenterAccess(true),
			),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterFAQ
func (t TrustCenterFAQ) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			immutable:  true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			name:       "notes",
			edgeSchema: Note{},
		}),
	}
}

// Modules this schema has access to
func (TrustCenterFAQ) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Policy of the TrustCenterFAQ
func (TrustCenterFAQ) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithOnMutationRules(ent.OpCreate,
			policy.CheckCreateAccess(),
		),
		policy.WithMutationRules(
			rule.AllowIfTrustCenterEditor(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterFAQMutation](),
		),
	)
}

// Annotations of the TrustCenterFAQ
func (TrustCenterFAQ) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}

// Interceptors of the TrustCenterFAQ
func (TrustCenterFAQ) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
