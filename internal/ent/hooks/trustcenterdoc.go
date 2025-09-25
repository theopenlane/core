package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/fgax"
)

var (
	errMissingFileID = errors.New("missing file id")
)

func HookTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center doc hook")

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)

			_, mutationSetsFileID := m.FileID()

			// check if the operation is a create operation
			if m.Op() == ent.OpCreate && len(fileIDs) < 1 && !mutationSetsFileID {
				return nil, objects.ErrNoFilesUploaded
			}

			if len(fileIDs) > 0 {
				var err error
				ctx, err = checkTrustCenterDocFile(ctx, m)
				if err != nil {
					return nil, err
				}
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			trustCenterDoc, ok := v.(*generated.TrustCenterDoc)
			if !ok {
				return v, nil
			}

			// TODO: this will happen once watermarking is enabled, and we do this async.
			// will need to re-write the file viewer logic once this happens
			if trustCenterDoc.FileID == nil {
				return nil, errMissingFileID
			}

			if m.Op() == ent.OpUpdateOne && (len(fileIDs) == 0 || mutationSetsFileID) {
				err = updateTrustCenterDocVisibility(ctx, m, *trustCenterDoc.FileID, trustCenterDoc.ID)
				return v, err
			}

			if m.Op() != ent.OpCreate {
				return v, nil
			}

			if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityNotVisible {
				return v, nil
			}
			tuples := fgax.CreateWildcardViewerTuple(trustCenterDoc.ID, "trust_center_doc")

			if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
				tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.FileID, generated.TypeFile)...)
			}

			if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
				return nil, err
			}

			return trustCenterDoc, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// updateTrustCenterDocVisibility updates fga tuples based on the visibility of the trust center doc
func updateTrustCenterDocVisibility(ctx context.Context, m *generated.TrustCenterDocMutation, fileID string, docID string) error {
	// 1. Check if the visibility of the document has changed
	visibility, visibilityChanged := m.Visibility()
	if !visibilityChanged {
		// No visibility change, nothing to do
		return nil
	}

	// Get the old visibility to compare
	oldVisibility, err := m.OldVisibility(ctx)
	if err != nil {
		return err
	}

	// If visibility hasn't actually changed, nothing to do
	if oldVisibility == visibility {
		return nil
	}

	writes := []fgax.TupleKey{}
	deletes := []fgax.TupleKey{}

	// 2. If the visibility has changed to public, add the wildcard tuples
	if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
		writes = append(writes, fgax.CreateWildcardViewerTuple(fileID, generated.TypeFile)...)
	}

	// 3. If the visibility has changed from public to not visible or protected, remove the wildcard tuples
	if oldVisibility == enums.TrustCenterDocumentVisibilityPubliclyVisible &&
		(visibility == enums.TrustCenterDocumentVisibilityNotVisible || visibility == enums.TrustCenterDocumentVisibilityProtected) {
		deletes = append(deletes, fgax.CreateWildcardViewerTuple(fileID, generated.TypeFile)...)
	}

	// 4. If the visibility changed from not visible -> protected or public, add the wildcard viewer tuples
	if oldVisibility == enums.TrustCenterDocumentVisibilityNotVisible {
		writes = append(writes, fgax.CreateWildcardViewerTuple(docID, "trust_center_doc")...)
	}

	// 5. If the visibility changed from protected or public -> not visible, remove the wildcard viewer tuples
	if visibility == enums.TrustCenterDocumentVisibilityNotVisible {
		deletes = append(deletes, fgax.CreateWildcardViewerTuple(docID, "trust_center_doc")...)
	}

	// Apply the tuple changes if any
	if len(writes) > 0 || len(deletes) > 0 {
		if _, err := m.Authz.WriteTupleKeys(ctx, writes, deletes); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msg("failed to update file access permissions")
			return err
		}
	}
	return nil
}

// checkTrustCenterDocFile checks if a trust center doc file is provided and sets the local file ID
func checkTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, error) {
	dockey := "trustCenterDocFile"

	// get the file from the context, if it exists
	docFile, _ := objects.FilesFromContextWithKey(ctx, dockey)
	if docFile == nil {
		return ctx, nil
	}

	// this should always be true, but check just in case
	if docFile[0].FieldName == dockey {
		// we should only have one file
		if len(docFile) > 1 {
			return ctx, ErrNotSingularUpload
		}
		m.SetOriginalFileID(docFile[0].ID)

		// TODO: once the watermarking job is implemented, we will not set the file ID here, and instead kick off a riverboat job to watermark the doc
		// If watermarking is disabled, we will continue to just set this file ID
		m.SetFileID(docFile[0].ID)

		docFile[0].Parent.ID, _ = m.ID()
		docFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, dockey, docFile[0])
	}

	return ctx, nil
}
