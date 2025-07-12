package schema

// TrustCenterCompliance represents compliance with a framework, and associates
// it with the organization's trust center When implemented, this will have a
// pointer to a "program" object and its "standard" framework

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
	return []ent.Field{}
}

// Mixin of the TrustCenterCompliance
func (t TrustCenterCompliance) Mixin() []ent.Mixin {
	return mixinConfig{}.getMixins()
}

// Edges of the TrustCenterCompliance
func (t TrustCenterCompliance) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Hooks of the TrustCenterCompliance
func (TrustCenterCompliance) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the TrustCenterCompliance
func (TrustCenterCompliance) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(),
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(),
		),
	)
}

// Indexes of the TrustCenterCompliance
func (TrustCenterCompliance) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterCompliance
func (TrustCenterCompliance) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Interceptors of the TrustCenterCompliance
func (TrustCenterCompliance) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
