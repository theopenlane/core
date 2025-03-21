package mixin

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// SystemOwnedMixin implements the revision pattern for schemas.
type SystemOwnedMixin struct {
	mixin.Schema
}

// Fields of the SystemOwnedMixin.
func (SystemOwnedMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("system_owned").
			Optional().
			Default(false).
			Annotations(
				// the field is automatically set to true if the user is a system admin
				// do not allow this field to be set in the mutation manually
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Immutable(). // don't allow this to be changed after creation, a new record must be created
			Comment("indicates if the record is owned by the the openlane system and not by an organization"),
	}
}

// Hooks of the SystemOwnedMixin.
func (d SystemOwnedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		HookSystemOwned(),
	}
}

// SystemOwnedMutation is an interface for interacting with the system_owned field in mutations
// it will add the system_owned_field and will automatically set the field to true if the user is a system admin
type SystemOwnedMutation interface {
	SetSystemOwned(bool)
}

// HookSystemOwned will automatically set the system_owned field to true if the user is a system admin
func HookSystemOwned() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			admin, err := rule.CheckIsSystemAdmin(ctx, m)
			if err != nil {
				log.Error().Err(err).Msg("unable to check if user is system admin, skipping setting system owned")
			}

			if admin {
				mut, ok := m.(SystemOwnedMutation)
				if ok && mut != nil {
					mut.SetSystemOwned(true)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}
