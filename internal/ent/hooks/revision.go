package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/slateparser"
)

// MutationWithRevision is an interface that defines the methods
// required for a mutation to be able to handle revisions
// It includes methods for getting and setting the revision
type MutationWithRevision interface {
	Revision() (string, bool)
	RevisionCleared() bool
	OldRevision(ctx context.Context) (string, error)
	SetRevision(s string)
	OldField(ctx context.Context, name string) (ent.Value, error)

	utils.GenericMutation
}

// HookRevisionUpdate is a hook that runs on update mutations
// to handle the revision of an object
// It checks if the revision is set, and if not, it retrieves the current revision from the database
// and bumps the patch version if just metadata was updated, bumps minor for details or details_json updates
// If the revision is cleared, it sets the revision to the default value
func HookRevisionUpdate() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			mut := m.(MutationWithRevision)

			if mut.RevisionCleared() {
				// if the revision is cleared, set it to the default
				mut.SetRevision(models.DefaultRevision)

				return next.Mutate(ctx, m)
			}

			// set the new revision
			if err := SetNewRevision(ctx, mut); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, mut.(ent.Mutation))
		})
	},
		hook.HasOp(ent.OpUpdateOne),
	)
}

// SetNewRevision sets the new revision for a mutation based on the current revision and the revision bump
// If the revision is set, it does nothing
// If the revision is not set, it retrieves the current revision from the database and bumps the version based on the revision bump
// If there is no revision bump set, it bumps the patch version
//
// TODO: known read-modify-write race — OldRevision + SetRevision is not serialized against
// concurrent updaters on the same row, so two parallel writes can compute the same vN+1 and
// the audit log loses one of them. The proper fix is to either (a) run this hook inside a
// transaction with SELECT ... FOR UPDATE on the target row before reading OldRevision, or
// (b) push the bump into SQL via "WHERE revision = $old" with retry on row-count == 0.
// Both require infrastructure changes beyond this hook. Track in a dedicated PR.
func SetNewRevision(ctx context.Context, mut MutationWithRevision) error {
	revision, ok := mut.Revision()

	currentRevision, err := mut.OldRevision(ctx)
	if err != nil {
		return err
	}

	// if the revision is set and the old and new don't match, return - user manually set
	if ok && revision != currentRevision {
		return nil
	}

	revisionBump, ok := models.VersionBumpFromRequestContext(ctx)
	if !ok {
		revisionBump = &models.Patch
		// EXTERNAL_REFERENCE docs (e.g. uploaded Word files) treat the file itself as
		// the source of truth, so a new upload — not a details edit — drives the bump.
		if doc, ok := mut.(documentMutation); ok && managementModeFor(ctx, doc) == enums.DocumentManagementModeExternalReference {
			if fileChanged(ctx, doc) {
				revisionBump = &models.Minor
			}
		} else if detailsUpdated(ctx, mut) {
			revisionBump = &models.Minor
		}
	}

	var newVersion string

	logx.FromContext(ctx).Debug().Str("currentRevision", currentRevision).
		Str("revisionBump", revisionBump.String()).
		Msg("bumping revision")

	switch *revisionBump {
	case models.Major:
		newVersion, err = models.BumpMajor(currentRevision)
	case models.Minor:
		newVersion, err = models.BumpMinor(currentRevision)
	case models.PreRelease:
		newVersion, err = models.SetPreRelease(currentRevision)
	default:
		newVersion, err = models.BumpPatch(currentRevision)
	}

	if err != nil {
		return err
	}

	// set the revision to the new revision
	mut.SetRevision(newVersion)

	return nil
}

const (
	detailsJSONFieldName = "details_json"
	detailsFieldName     = "details"
)

// documentMutation is the subset of revision-bearing mutations that carry a document file
// and a management mode. Action plans, internal policies, and procedures satisfy it; other
// revision-bearing mutations (standards, templates, control objectives) do not.
type documentMutation interface {
	ManagementMode() (enums.DocumentManagementMode, bool)
	OldManagementMode(ctx context.Context) (enums.DocumentManagementMode, error)
	FileID() (r string, exists bool)
	OldFileID(ctx context.Context) (v *string, err error)
	FileIDCleared() bool
}

// managementModeFor returns the mode that will be in effect after this mutation.
// Defaults to OPENLANE_MANAGED so pre-existing rows (NULL column) and create-path
// mutations behave as managed documents.
func managementModeFor(ctx context.Context, mut documentMutation) enums.DocumentManagementMode {
	if v, ok := mut.ManagementMode(); ok && v.IsValid() {
		return v
	}

	if v, err := mut.OldManagementMode(ctx); err == nil && v.IsValid() {
		return v
	} else if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("could not read old management_mode; defaulting to OPENLANE_MANAGED")
	}

	return enums.DocumentManagementModeOpenlaneManaged
}

// fileChanged reports whether this mutation replaces, sets, or clears the document's file.
// Used to drive a minor revision bump for EXTERNAL_REFERENCE documents where the file is the
// source of truth.
func fileChanged(ctx context.Context, mut documentMutation) bool {
	newID, set := mut.FileID()
	cleared := mut.FileIDCleared()
	if !set && !cleared {
		return false
	}

	oldID, err := mut.OldFileID(ctx)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("could not read old file_id; treating as file change")
		return true
	}

	oldStr := ""
	if oldID != nil {
		oldStr = *oldID
	}

	if cleared {
		return oldStr != ""
	}

	return oldStr != newID
}

// detailsUpdated checks if the details were updated on a mutation, if so
// this will return true, otherwise returns false.
// this is always an updateOne mutation that we are concerned about, so we can simplify
// any logic related to create/delete/update many
func detailsUpdated(ctx context.Context, m MutationWithRevision) bool {
	fields := m.Fields()

	// check of the changed fields contains details json
	// if so, if only comments were added do not bump minor
	// if more than comments were added, bump minor
	if lo.Contains(fields, detailsJSONFieldName) {
		// get the old details json value from the mutation
		oldDetailsJSON, _ := m.OldField(ctx, detailsJSONFieldName)
		newDetailsJSON, _ := m.Field(detailsJSONFieldName)

		oldDetailsTyped, _ := oldDetailsJSON.([]any)
		newDetailsTyped, _ := newDetailsJSON.([]any)

		return !slateparser.NoDetailsChanged(oldDetailsTyped, newDetailsTyped)
	}

	// if details json is not set, fallback to check details
	// this will not need to check comments as this would be be an API update
	// UI will always set details_json.
	if lo.Contains(fields, detailsFieldName) {
		// get the old details json value from the mutation
		oldDetails, oldErr := m.OldField(ctx, detailsFieldName)
		newDetails, newOk := m.Field(detailsFieldName)
		if !newOk || (oldErr != nil && !newOk) {
			return false
		}

		if oldErr != nil && newOk {
			return true
		}

		oldDetailsString, oldOk := oldDetails.(string)
		newDetailsString, newOk := newDetails.(string)
		if !newOk || (!oldOk && !newOk) {
			return false
		}

		// if the old is not ok,  but new is, that indicates the new value is set and old value is not
		// so this is an update
		if !oldOk && newOk {
			return true
		}

		if oldDetailsString == newDetailsString {
			return false
		}

		return true
	}

	return false
}
