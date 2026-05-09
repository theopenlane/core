package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/samber/lo"
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
		// derive based on if there were meaningful updates to the details of the document
		revisionBump = &models.Patch
		if detailsUpdated(ctx, mut) {
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
