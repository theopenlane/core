package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
	}
}

// Mixin of the TrustCenter
func (t TrustCenter) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t),
		},
	}.getMixins()
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
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema:    t,
			t:             TrustCenterSetting.Type,
			cascadeDelete: "TrustCenter",
			name:          "setting",
		}),
	}
}

// Hooks of the TrustCenter
func (TrustCenter) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenter(),
	}
}

// Policy of the TrustCenter
func (TrustCenter) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
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
		index.Fields(ownerFieldName).
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}
