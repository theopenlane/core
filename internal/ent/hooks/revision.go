package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// MutationWithRevision is an interface that defines the methods
// required for a mutation to be able to handle revisions
// It includes methods for getting and setting the revision
type MutationWithRevision interface {
	Revision() (string, bool)
	RevisionCleared() bool
	OldRevision(ctx context.Context) (string, error)
	SetRevision(s string)

	utils.GenericMutation
}

// HookRevisionUpdate is a hook that runs on update mutations
// to handle the revision of an object
// It checks if the revision is set, and if not, it retrieves the current revision from the database
// and bumps the patch version
// If the revision is cleared, it sets the revision to the default value
func HookRevisionUpdate() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
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
	// if the revision is set, continue
	revision, ok := mut.Revision()
	if ok && revision != "" {
		// revision is already set, do nothing
		return nil
	}

	currentRevision, err := mut.OldRevision(ctx)
	if err != nil {
		return err
	}

	revisionBump, ok := models.VersionBumpFromRequestContext(ctx)
	if !ok {
		revisionBump = &models.Patch
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
