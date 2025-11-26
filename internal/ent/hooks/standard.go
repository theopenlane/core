package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

var (
	// ErrPublicStandardCannotBeDeleted defines an error that denotes a public standard cannot be
	// deleted once made public
	ErrPublicStandardCannotBeDeleted = errors.New("public standard not allowed to be deleted")
)

// HookStandardDelete cascades the deletion of all controls for a system-owned standard connected
// as long as the standard is not public. This is to prevent the deletion of a standard that is
// actively used by an organization.
func HookStandardDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.StandardFunc(func(ctx context.Context, m *generated.StandardMutation) (generated.Value, error) {
			// only run on delete operations
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			// only system admins edit system-owned standards
			// check early to avoid unnecessary database queries to both
			// the database and the authz service
			if !auth.IsSystemAdminFromContext(ctx) {
				return next.Mutate(ctx, m)
			}

			id, _ := m.ID()

			var err error

			// use the same context we use on edge_cleanup to make sure everything is cleaned up properly
			ctx = contextx.With(privacy.DecisionContext(ctx, privacy.Allowf("cleanup standard control edges")), entfga.DeleteTuplesFirstKey{})

			// get the standard, we only need the systemOwned and isPublic fields
			retrievedStandard, err := m.Client().Standard.Query().
				Where(standard.ID(id)).
				Select(standard.FieldSystemOwned, standard.FieldIsPublic).
				Only(ctx)
			if err != nil {
				return nil, err
			}

			// if the standard is not system-owned, we don't need to do anything
			// we don't want to cascade on org owned standards because they might be used
			// by an organization and they don't care about the parent standard
			if !retrievedStandard.SystemOwned {
				return next.Mutate(ctx, m)
			}

			// prevent accidental deletion of public standards, and require a system admin
			// to flip it to not public before they can delete it
			if retrievedStandard.IsPublic {
				return nil, ErrPublicStandardCannotBeDeleted
			}

			// remove standard_id mapping from org owned controls
			// this uses the same allow context, as above, which will allow the
			// control to be updated to clear the standard id field by the system admin
			err = m.Client().Control.Update().ClearStandardID().
				Where(
					control.And(
						control.StandardID(id),
						control.OwnerIDNotNil(),
					),
				).Exec(ctx)
			if err != nil {
				return nil, err
			}

			// delete all controls not linked to an org (system-owned controls)
			_, err = m.Client().Control.Delete().Where(
				control.And(
					control.StandardID(id),
					control.OwnerIDIsNil(),
				),
			).Exec(ctx)
			if err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne|ent.OpUpdate)
}

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

func HookStandardFileUpload() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.StandardFunc(func(ctx context.Context, m *generated.StandardMutation) (generated.Value, error) {
			// check for uploaded files (e.g. logo image)
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkStandardLogoFile(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	},
		hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne),
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
func AddOrDeletePublicStandardTuple(ctx context.Context, m *generated.StandardMutation) (bool, bool, error) {
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

	return false, false, nil
}

// standardTupleOnCreate adds the tuple if both systemOwned and isPublic are true
func standardTupleOnCreate(m *generated.StandardMutation) (bool, bool, error) {
	systemOwned, ok := m.SystemOwned()
	if !ok {
		return false, false, nil
	}

	isPublic, ok := m.IsPublic()
	if !ok {
		return false, false, nil
	}

	if systemOwned && isPublic {
		return true, false, nil
	}

	return false, false, nil
}

// standardTupleOnUpdateOne deletes the tuple if it's a soft delete or if systemOwned or isPublic fields have changed. Adds the tuple if both fields are true
func standardTupleOnUpdateOne(ctx context.Context, m *generated.StandardMutation) (add, remove bool, err error) {
	var (
		oldPublic          bool
		oldSystemOwned     bool
		publicCleared      bool
		systemOwnedCleared bool
	)

	// if its a soft delete, delete the tuples
	if isDeleteOp(ctx, m) {
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
		remove = true
	}

	// delete logic if the systemOwned or isPublic fields have changed from true to false
	if !systemOwned && systemOwnedOK && oldSystemOwned || !public && publicOK && oldPublic {
		remove = true
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
func standardTupleOneUpdate(ctx context.Context, m *generated.StandardMutation) (bool, bool, error) {
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
		updatedIDs := getMutationIDs(ctx, m)
		if len(updatedIDs) == 0 {
			return false, false, nil
		}

		// check if the systemOwned or isPublic fields have changed
		stds, err := m.Client().Standard.Query().
			Select(standard.FieldIsPublic, standard.FieldSystemOwned).
			Where(standard.IDIn(updatedIDs...)).
			All(ctx)
		if err != nil {
			return false, false, err
		}

		for _, std := range stds {
			if std.SystemOwned && std.IsPublic {
				return true, false, nil
			}
		}

	}

	return false, false, nil
}

func checkStandardLogoFile(ctx context.Context, m *generated.StandardMutation) (context.Context, error) {
	// logoKey := "logoFile"

	// logoFiles, _ := pkgobjects.FilesFromContextWithKey(ctx, logoKey)
	// if len(logoFiles) == 0 {
	// 	return ctx, nil
	// }

	// if len(logoFiles) > 1 {
	// 	return ctx, ErrTooManyLogoFiles
	// }

	// m.SetLogoFileID(logoFiles[0].ID)

	// adapter := pkgobjects.NewGenericMutationAdapter(m,
	// 	func(mut *generated.StandardMutation) (string, bool) { return mut.ID() },
	// 	func(mut *generated.StandardMutation) string { return mut.Type() },
	// )

	// return pkgobjects.ProcessFilesForMutation(ctx, adapter, logoKey)

	return ctx, nil
}
