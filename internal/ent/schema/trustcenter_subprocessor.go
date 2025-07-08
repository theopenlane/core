package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
	return []ent.Field{}
}

// Mixin of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Mixin() []ent.Mixin {
	return mixinConfig{}.getMixins()
}

// Edges of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Hooks of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(),
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(),
		),
	)
}

// Indexes of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Interceptors of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
