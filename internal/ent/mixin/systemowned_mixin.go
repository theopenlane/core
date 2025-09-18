package mixin

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

// SystemOwnedMixin implements the revision pattern for schemas.
type SystemOwnedMixin struct {
	mixin.Schema

	// AdditionalAdminOnlyFields can be set to add additional fields that are only
	// available to system admins
	AdditionalAdminOnlyFields []string
}

func NewSystemOwnedMixin(additionalFields ...string) SystemOwnedMixin {
	return SystemOwnedMixin{
		AdditionalAdminOnlyFields: additionalFields,
	}
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
		field.String("internal_notes").
			Optional().
			Comment("internal notes about the object creation, this field is only available to system admins").
			Nillable(),
		field.String("system_internal_id").
			Optional().
			Comment("an internal identifier for the mapping, this field is only available to system admins").
			Nillable(),
	}
}

// Hooks of the SystemOwnedMixin.
func (d SystemOwnedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		HookSystemOwnedCreate(),
	}
}

func (SystemOwnedMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the SystemOwnedMixin
func (s SystemOwnedMixin) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.AllowMutationIfSystemAdmin(),
			SystemOwnedSchema(s.AdditionalAdminOnlyFields...),
		},
	}
}

// SystemOwnedMutation is an interface for interacting with the system_owned field in mutations
// it will add the system_owned_field and will automatically set the field to true if the user is a system admin
type SystemOwnedMutation interface {
	utils.GenericMutation

	Field(name string) (ent.Value, bool)
	FieldCleared(name string) bool
	SystemOwned() (bool, bool)
	SetSystemOwned(bool)
	OldSystemOwned(context.Context) (bool, error)
	SystemInternalID() (string, bool)
	SetSystemInternalID(string)
	InternalNotes() (string, bool)
	SetInternalNotes(string)
}

// HookSystemOwnedCreate will automatically set the system_owned field to true if the user is a system admin
func HookSystemOwnedCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			admin, err := rule.CheckIsSystemAdminWithContext(ctx)
			if err != nil {
				zerolog.Ctx(ctx).Error().Err(err).Msg("unable to check if user is system admin, skipping setting system owned")

				return next.Mutate(ctx, m)
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

// InterceptorSystemFields handles returning internal only fields for system owned schemas
func InterceptorSystemFields() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		admin, err := rule.CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if admin {
			return nil
		}

		// if not a system admin, do not return system owned fields
		return nil
	})
}

func SystemOwnedSchema(additionalFields ...string) privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		mut, ok := m.(SystemOwnedMutation)
		if !ok || mut == nil {
			return privacy.Skipf("not a system owned mutation")
		}

		systemOwned, _ := mut.SystemOwned()
		internalID, _ := mut.SystemInternalID()
		internalNotes, _ := mut.InternalNotes()

		hasAdditionalAdminField := false
		// check any additional fields that should be admin only
		for _, field := range additionalFields {
			value, ok := mut.Field(field)
			if ok && value != nil && value != "" && value != false {
				hasAdditionalAdminField = true
				break
			}

			if mut.FieldCleared(field) {
				hasAdditionalAdminField = true
				break
			}
		}

		hasAdminField := systemOwned || hasAdditionalAdminField || internalID != "" || internalNotes != ""

		admin, err := rule.CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if admin {
			return privacy.Allow
		}

		// if the field was not in the mutation, check the database
		if !hasAdminField {
			switch m.Op() {
			case ent.OpCreate:
				// on create check if system owned is being set, if not continue
				return privacy.Skipf("no system owned field set")
			default:
				// on update, update one, delete, delete one, always check
				// to ensure the system owned field is set
				ids, err := mut.IDs(ctx)
				if err != nil {
					return err
				}

				systemOwned, err = queryForSystemOwned(ctx, mut, ids)
				if err != nil {
					return err
				}

				if systemOwned {
					return generated.ErrPermissionDenied
				}
			}
		}

		if hasAdminField {
			zerolog.Ctx(ctx).Debug().Msg("user attempted to set system owned field without being a system admin")

			return fmt.Errorf("%w: %w", hooks.ErrInvalidInput, rule.ErrAdminOnlyField)
		}

		return privacy.Skip
	})
}

func queryForSystemOwned(ctx context.Context, m SystemOwnedMutation, ids []string) (bool, error) {
	// if no ids, return false and continue to the next rule
	// this would happen if the object being mutated does not exist
	if len(ids) == 0 {
		return false, nil
	}

	table := strcase.SnakeCase(pluralize.NewClient().Plural(m.Type()))
	query := "SELECT system_owned FROM " + table + " WHERE id in ($1)"

	var rows sql.Rows
	if err := m.Client().Driver().Query(ctx, query, toAnySlice(ids), &rows); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("failed to check for object system owned status")

		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		var systemOwned bool
		if err := rows.Scan(&systemOwned); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to scan system owned field")

			return false, err
		}

		if systemOwned {
			return true, nil
		}
	}

	return false, nil
}

func toAnySlice(input []string) []any {
	output := make([]any, len(input))
	for i, v := range input {
		output[i] = v
	}

	return output
}
