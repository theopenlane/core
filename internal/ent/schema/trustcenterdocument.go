package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
	return []ent.Field{}
}

// Mixin of the TrustCenterDoc
func (t TrustCenterDoc) Mixin() []ent.Mixin {
	return mixinConfig{}.getMixins(t)
}

// Edges of the TrustCenterDoc
func (t TrustCenterDoc) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Hooks of the TrustCenterDoc
func (TrustCenterDoc) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the TrustCenterDoc
func (TrustCenterDoc) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(),
		),
	)
}

// Indexes of the TrustCenterDoc
func (TrustCenterDoc) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterDoc
func (TrustCenterDoc) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Interceptors of the TrustCenterDoc
func (TrustCenterDoc) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
