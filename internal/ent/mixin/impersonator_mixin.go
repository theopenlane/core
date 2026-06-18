package mixin

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hooks"
)

// ImpersonatorMixin records the real actor behind an impersonation session (such as an Openlane
// support session) on each mutated record. created_by/updated_by continue to reflect the impersonated
// identity, while updated_by_impersonator names the person acting through it so changes stay traceable
type ImpersonatorMixin struct {
	mixin.Schema
}

// Fields of the ImpersonatorMixin
func (ImpersonatorMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("updated_by_impersonator").
			Comment("the real user acting through an impersonation session when the record was last mutated, if any").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			),
	}
}

// Hooks of the ImpersonatorMixin
func (ImpersonatorMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookImpersonatorAttribution(),
	}
}
