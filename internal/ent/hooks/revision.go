package hooks

import (
	"context"
	"encoding/json"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/models"
)

// MutationWithRevision is an interface that defines the methods
// required for a mutation to be able to handle revisions
// It includes methods for getting and setting the revision
type MutationWithRevision interface {
	Revision() (*models.SemverVersion, bool)
	RevisionCleared() bool
	SetRevision(mv *models.SemverVersion)

	GenericMutation
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
				mut.SetRevision(&models.SemverVersion{
					Patch: 1,
				})

				return next.Mutate(ctx, m)
			}

			// if the revision is set, continue
			_, ok := mut.Revision()
			if ok {
				return next.Mutate(ctx, m)
			}

			currentRevision, err := getRevisionFromDatabase(ctx, mut)
			if err != nil {
				return nil, err
			}

			// bump the patch version
			currentRevision.BumpPatch()

			// set the revision to the current revision
			mut.SetRevision(currentRevision)

			return next.Mutate(ctx, mut.(ent.Mutation))
		})
	},
		hook.HasOp(ent.OpUpdateOne),
	)
}

// getRevisionFromDatabase retrieves the current revision of an object from the database.
// It takes a context and a MutationWithRevision as parameters and returns a SemverVersion model and an error
// if the revision cannot be found
func getRevisionFromDatabase(ctx context.Context, m MutationWithRevision) (*models.SemverVersion, error) {
	var revision *models.SemverVersion

	// get current revision from the database
	objectType := strings.ToLower(m.Type())

	// get updated objectIDs
	objectIDs, err := m.IDs(ctx)
	if err != nil {
		return nil, err
	}

	// table is always the pluralized version of the object type
	table := pluralize.NewClient().Plural(objectType)
	query := "SELECT revision FROM " + table + " WHERE id in ($1)"

	var rows sql.Rows

	if err := generated.FromContext(ctx).Driver().Query(ctx, query, []any{strings.Join(objectIDs, ",")}, &rows); err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		var revisionValue []byte

		if err := rows.Scan(&revisionValue); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(revisionValue, &revision); err != nil {
			return nil, err
		}
	}

	return revision, nil
}
