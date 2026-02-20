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
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
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
		field.String("note_id").
			Comment("ID of the note containing the FAQ question and answer").
			Immutable().
			NotEmpty(),
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Immutable().
			Optional(),
		field.String("reference_link").
			Comment("optional reference link for the FAQ").
			Validate(validator.ValidateURL()).
			Optional(),
		field.Int("display_order").
			Comment("display order of the FAQ").
			Default(0).
			Optional(),
	}
}

// Mixin of the TrustCenterFAQ
func (t TrustCenterFAQ) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterFAQ](t,
				withParents(TrustCenter{}),
				withAllowAnonymousTrustCenterAccess(true),
			),
			newCustomEnumMixin(t),
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
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Note{},
			immutable:  true,
			field:      "note_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Organization{}.Name()),
			},
		}),
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

// Indexes of the TrustCenterFAQ
func (TrustCenterFAQ) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("note_id", "trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (TrustCenterFAQ) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Annotations of the TrustCenterFAQ
func (TrustCenterFAQ) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
		entx.NewExportable(),
		entgql.QueryField("trustCenterFAQs"),
		entsql.Annotation{Table: "trust_center_faqs"},
	}
}

// Interceptors of the TrustCenterFAQ
func (TrustCenterFAQ) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
