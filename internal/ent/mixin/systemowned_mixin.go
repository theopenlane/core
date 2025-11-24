package mixin

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/internal/graphapi/directives"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
)

const (
	// SystemOwnedMixinName is the name of the SystemOwnedMixin
	SystemOwnedMixinName = "SystemOwnedMixin"
)

// SystemOwnedMixin implements the revision pattern for schemas.
type SystemOwnedMixin struct {
	mixin.Schema
}

// NewSystemOwnedMixin creates a new SystemOwnedMixin with the given options.
// The options can be used to customize the behavior of the mixin, however, there are currently no options.
func NewSystemOwnedMixin(opts ...SystemOwnedMixinOption) SystemOwnedMixin {
	m := SystemOwnedMixin{}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// SystemOwnedMixinOption is a function that configures the SystemOwnedMixin
type SystemOwnedMixinOption func(*SystemOwnedMixin)

// Name of the SystemOwnedMixin
func (SystemOwnedMixin) Name() string {
	return "SystemOwnedMixin"
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
			Annotations(
				directives.HiddenDirectiveAnnotation,
			).
			Nillable(),
		field.String("system_internal_id").
			Optional().
			Comment("an internal identifier for the mapping, this field is only available to system admins").
			Annotations(
				directives.HiddenDirectiveAnnotation,
			).
			Nillable(),
	}
}

// Hooks of the SystemOwnedMixin.
func (d SystemOwnedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		HookSystemOwnedCreate(),
	}
}

// Policy of the SystemOwnedMixin
func (d SystemOwnedMixin) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.AllowMutationIfSystemAdmin(),
			SystemOwnedSchema(),
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
	ClearSystemInternalID()
	SetSystemInternalID(string)
	InternalNotes() (string, bool)
	ClearInternalNotes()
	SetInternalNotes(string)
	OwnerID() (string, bool)
	SetOwnerID(string)
}

// OrgOwnedMutation is an interface for interacting with the owner_id field in mutations
type OrgOwnedMutation interface {
	utils.GenericMutation

	OwnerID() (string, bool)
	SetOwnerID(string)
}

// HookSystemOwnedCreate will automatically set the system_owned field to true if the user is a system admin
// and ensure there is an owner id when creating not system owned objects
func HookSystemOwnedCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			admin, err := rule.CheckIsSystemAdminWithContext(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("unable to check if user is system admin, skipping setting system owned")

				return next.Mutate(ctx, m)
			}

			mut, ok := m.(SystemOwnedMutation)
			if !ok && mut == nil {
				return next.Mutate(ctx, m)
			}

			if admin {
				mut.SetSystemOwned(true)

				return next.Mutate(ctx, m)
			}

			// if its not a system admin, ensure the system owned field is false
			mut.SetSystemOwned(false)

			// ensure there is an owner id set for non system owned objects
			orgMut, ok := m.(OrgOwnedMutation)
			if !ok && orgMut == nil {
				return next.Mutate(ctx, m)
			}

			ownerID, ok := orgMut.OwnerID()
			if !ok || ownerID == "" {
				logx.FromContext(ctx).Debug().Msg("non system admin creating object without owner ID, attempting to set")

				orgID, err := auth.GetOrganizationIDFromContext(ctx)
				if err != nil || orgID == "" {
					logx.FromContext(ctx).Error().Err(err).Msg("unable to get organization ID from context for non system admin creating object")

					return nil, generated.ErrPermissionDenied
				}

				mut.SetOwnerID(orgID)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// SystemOwnedSchema is a privacy rule that checks if the object is system owned
// and if the user is a system admin
// For create operations, since the field is automatically set, we skip the check
// For update operations, the rule checks if the existing object is system owned
// and denys if it is and the user is not a system admin
func SystemOwnedSchema() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		// on create check continue, the field is automatically set based on user role
		if m.Op() == ent.OpCreate {
			return privacy.Skip
		}

		mut, ok := m.(SystemOwnedMutation)
		if !ok || mut == nil {
			return privacy.Skipf("not a system owned mutation")
		}

		admin, err := rule.CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return err
		}

		if admin {
			return privacy.Allow
		}

		systemOwned, _ := mut.SystemOwned()
		if systemOwned {
			logx.FromContext(ctx).Warn().Msg("attempt to modify system owned object by non system admin")

			return generated.ErrPermissionDenied
		}

		// if the field was not in the mutation, check the database
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
			logx.FromContext(ctx).Warn().Msg("attempt to modify system owned object by non system admin")

			return generated.ErrPermissionDenied
		}

		return privacy.Skip
	})
}

// queryForSystemOwned checks the database to see if any of the objects are system owned
func queryForSystemOwned(ctx context.Context, m SystemOwnedMutation, ids []string) (bool, error) {
	// if no ids, return false and continue to the next rule
	// this would happen if the object being mutated does not exist
	if len(ids) == 0 {
		return false, nil
	}

	table := strcase.SnakeCase(pluralize.NewClient().Plural(m.Type()))
	query := "SELECT system_owned FROM " + table + " WHERE id in ($1)"

	var rows sql.Rows
	if err := m.Client().Driver().Query(ctx, query, lo.ToAnySlice(ids), &rows); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to check for object system owned status")

		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		var systemOwned bool
		if err := rows.Scan(&systemOwned); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to scan system owned field")

			return false, err
		}

		if systemOwned {
			return true, nil
		}
	}

	return false, nil
}
