package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// Subprocessor holds the schema definition for the Subprocessor entity
type Subprocessor struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSubprocessor is the name of the Subprocessor schema.
const SchemaSubprocessor = "subprocessor"

// Name returns the name of the Subprocessor schema.
func (Subprocessor) Name() string {
	return SchemaSubprocessor
}

// GetType returns the type of the Subprocessor schema.
func (Subprocessor) GetType() any {
	return Subprocessor.Type
}

// PluralName returns the plural name of the Subprocessor schema.
func (Subprocessor) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSubprocessor)
}

// Fields of the Subprocessor
func (Subprocessor) Fields() []ent.Field {
	return []ent.Field{}
}

// Mixin of the Subprocessor
func (t Subprocessor) Mixin() []ent.Mixin {
	return mixinConfig{}.getMixins()
}

// Edges of the Subprocessor
func (t Subprocessor) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Hooks of the Subprocessor
func (Subprocessor) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the Subprocessor
func (Subprocessor) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysDenyRule(),
		),
		policy.WithMutationRules(
			privacy.AlwaysDenyRule(),
		),
	)
}

// Indexes of the Subprocessor
func (Subprocessor) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Subprocessor
func (Subprocessor) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

// Interceptors of the Subprocessor
func (Subprocessor) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
