package hooks

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

// HookEvidenceFiles runs on evidence mutations to check for uploaded files
func HookEvidenceFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.EvidenceFunc(func(ctx context.Context, m *generated.EvidenceMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				// validate creation date if only
				// - it is a create operation
				// - it was provided in an update operation
				creationDate, ok := m.CreationDate()
				op := m.Op()

				if op == ent.OpCreate && !ok {
					return nil, ErrZeroTimeNotAllowed
				}

				if ok && creationDate.After(time.Now()) {
					return nil, ErrFutureTimeNotAllowed
				}
			}

			// check for uploaded files (e.g. avatar image)
			fileIDs := objects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = checkEvidenceFiles(ctx, m)
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			if err := checkControlEdges(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// checkEvidenceFiles checks if a evidence files are provided and sets the local file ID(s)
func checkEvidenceFiles[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "evidenceFiles"

	// Create adapter for the existing mutation interface
	adapter := objects.NewGenericMutationAdapter(m,
		func(mut T) (string, bool) { return mut.ID() },
		func(mut T) string { return mut.Type() },
	)

	// Use the generic helper to process files
	return objects.ProcessFilesForMutation(ctx, adapter, key)
}
