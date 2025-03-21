package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
)

// HookStandardCreate sets default values on creation, such as setting the short name to the name if it's not provided
func HookStandardCreate() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.StandardFunc(func(ctx context.Context, m *generated.StandardMutation) (generated.Value, error) {
			shortName, ok := m.ShortName()
			if !ok || shortName == "" {
				// name is required on creation
				name, _ := m.Name()

				// if the short name is not set, set it to the name
				m.SetShortName(name)
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.HasOp(ent.OpCreate),
	)
}

// HookStandardPublicAccessTuples adds tuples for publicly available standards
// based on the system owned and isPublic fields; and deletes them when the fields are cleared.
// see AddOrDeleteStandardTuple for details on how the fields are checked and it's called functions
// for specifics on mutation types
func HookStandardPublicAccessTuples() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.StandardFunc(func(ctx context.Context, m *generated.StandardMutation) (generated.Value, error) {
			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return retVal, err
			}

			addTuple, deleteTuple, err := AddOrDeletePublicStandardTuple(ctx, m)
			if err != nil {
				return retVal, err
			}

			if !addTuple && !deleteTuple {
				return retVal, nil
			}

			writes := []fgax.TupleKey{}
			deletes := []fgax.TupleKey{}

			// get the IDs that were updated
			ids, err := GetObjectIDsFromMutation(ctx, m, retVal)
			if err != nil {
				return retVal, err
			}

			for _, id := range ids {
				if addTuple {
					writes = append(writes, fgax.CreateWildcardViewerTuple(id, generated.TypeStandard)...)
				}

				if deleteTuple {
					deletes = append(deletes, fgax.CreateWildcardViewerTuple(id, generated.TypeStandard)...)
				}
			}

			if len(writes) > 0 || len(deletes) > 0 {
				if _, err := m.Authz.WriteTupleKeys(ctx, writes, deletes); err != nil {
					return retVal, err
				}
			}

			return retVal, nil
		})
	},
		hook.HasOp(
			ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne),
	)
}

// AddOrDeletePublicStandardTuple determines whether to add or delete a standard tuple based on the mutation operation and field values.
//
// Parameters:
// - ctx: The context for the operation.
// - m: The StandardMutation containing the mutation details.
//
// Returns:
// - add: A boolean indicating whether to add the tuple.
// - delete: A boolean indicating whether to delete the tuple.
// - err: An error if any occurred during the operation.
//
// The function handles the following mutation operations:
// - OpCreate: Adds the tuple if both systemOwned and isPublic are true.
// - OpDelete, OpDeleteOne: Deletes the tuple.
// - OpUpdateOne: Deletes the tuple if it's a soft delete or if isPublic fields has changed. Adds the tuple if both fields are true.
// - OpUpdate: Deletes the tuple if isPublic field has been cleared. Adds the tuple if both fields are true.
func AddOrDeletePublicStandardTuple(ctx context.Context, m *generated.StandardMutation) (add, delete bool, err error) {
	switch m.Op() {
	case ent.OpCreate:
		return standardTupleOnCreate(m)
	case ent.OpDelete, ent.OpDeleteOne:
		return false, true, nil // on delete, delete the tuples
	case ent.OpUpdateOne:
		return standardTupleOnUpdateOne(ctx, m)
	case ent.OpUpdate:
		return standardTupleOneUpdate(ctx, m)
	}

	return
}

// standardTupleOnCreate adds the tuple if both systemOwned and isPublic are true
func standardTupleOnCreate(m *generated.StandardMutation) (add, delete bool, err error) {
	systemOwned, ok := m.SystemOwned()
	if !ok {
		return
	}

	isPublic, ok := m.IsPublic()
	if !ok {
		return
	}

	if systemOwned && isPublic {
		return true, false, nil
	}

	return
}

// standardTupleOnUpdateOne deletes the tuple if it's a soft delete or if systemOwned or isPublic fields have changed. Adds the tuple if both fields are true
func standardTupleOnUpdateOne(ctx context.Context, m *generated.StandardMutation) (add, delete bool, err error) {
	var (
		oldPublic          bool
		oldSystemOwned     bool
		publicCleared      bool
		systemOwnedCleared bool
	)

	// if its a soft delete, delete the tuples
	if entx.CheckIsSoftDelete(ctx) {
		return false, true, nil
	}
	// check if the systemOwned or isPublic fields have changed
	systemOwned, systemOwnedOK := m.SystemOwned()

	systemOwnedCleared = m.SystemOwnedCleared()
	if !systemOwnedCleared {
		oldSystemOwned, err = m.OldSystemOwned(ctx)
		if err != nil {
			return
		}
	}

	public, publicOK := m.IsPublic()

	publicCleared = m.IsPublicCleared()
	if !publicCleared {
		oldPublic, err = m.OldIsPublic(ctx)
		if err != nil {
			return
		}
	}

	// if either were cleared, delete the tuples
	if systemOwnedCleared || publicCleared {
		delete = true
	}

	// delete logic if the systemOwned or isPublic fields have changed from true to false
	if !systemOwned && systemOwnedOK && oldSystemOwned || !public && publicOK && oldPublic {
		delete = true
	}

	// add logic when both are set
	// no need to check ok because it will only be true if ok is also true
	if public && systemOwned {
		add = true
	}

	// if public was set but not system owned, but old system owned was OK
	if public && oldSystemOwned && !systemOwnedOK {
		add = true
	}

	// reverse for system owned
	if systemOwned && oldPublic && !publicOK {
		add = true
	}

	return
}

// standardTupleOneUpdate deletes the tuple if systemOwned or isPublic fields have been cleared. Adds the tuple if both fields are true
func standardTupleOneUpdate(ctx context.Context, m *generated.StandardMutation) (add, delete bool, err error) {
	var (
		publicCleared bool
	)

	shouldDelete := false
	// check if the systemOwned or isPublic fields have changed
	systemOwned, systemOwnedOK := m.SystemOwned()

	public, publicOK := m.IsPublic()
	publicCleared = m.IsPublicCleared()

	var oldPublic *bool
	if m.Op() == ent.OpUpdateOne {
		oldValue, err := m.OldIsPublic(ctx)
		if err != nil {
			return false, false, err
		}

		oldPublic = &oldValue
	}

	if publicCleared {
		public = false

		// if we took an action to clear the public field, we should delete the tuples
		shouldDelete = true
	}

	// if these are both true, add the tuple, and conditionally delete the tuple
	if systemOwned && public && (oldPublic == nil || public != *oldPublic) {
		return true, shouldDelete, nil
	}

	// if either of these we set in the mutation, we should delete the tuples
	if (!systemOwned && systemOwnedOK) || (!public && publicOK) {
		return false, true, nil
	}

	// see if we need to add the tuples because one of the two fields was set
	if systemOwned || public {
		var updatedIDs []string

		updatedIDs, err = m.IDs(ctx)
		if err != nil || len(updatedIDs) == 0 {
			return
		}

		// check if the systemOwned or isPublic fields have changed
		for _, id := range updatedIDs {
			var standard *generated.Standard

			standard, err = m.Client().Standard.Get(ctx, id)
			if err != nil {
				return
			}

			if standard.SystemOwned && standard.IsPublic {
				return true, false, nil
			}
		}
	}

	return
}
