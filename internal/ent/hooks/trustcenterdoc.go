package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"
)

var (
	errMissingFileID         = errors.New("missing file id")
	errCannotSetFileOnCreate = errors.New("cannot set file id on create")
)

// internalTrustCenterDocUpdateKey is used to mark internal update operations within hooks
type internalTrustCenterDocUpdateKey struct{}

func HookCreateTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("trust center doc create hook")
			// check for uploaded files (e.g. logo image)
			fileIDs, _ := objects.FilesFromContextWithKey(ctx, "watermarkedTrustCenterDocFile")

			_, mutationSetsFileID := m.FileID()
			_, mutationSetsOriginalFileID := m.OriginalFileID()

			if mutationSetsFileID || len(fileIDs) > 0 {
				return nil, errCannotSetFileOnCreate
			}

			var err error
			var fileUploaded bool
			ctx, fileUploaded, err = checkTrustCenterDocFile(ctx, m)
			if err != nil {
				return nil, err
			}

			if !mutationSetsOriginalFileID && !fileUploaded {
				return nil, objects.ErrNoFilesUploaded
			}

			watermarkingEnabled, watermarkingEnabledSet := m.WatermarkingEnabled()
			if !watermarkingEnabledSet || !watermarkingEnabled {
				origFileID, origFileIDSet := m.OriginalFileID()
				if !origFileIDSet {
					return nil, errMissingFileID
				}
				m.SetFileID(origFileID)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			trustCenterDoc, ok := v.(*generated.TrustCenterDoc)
			if !ok {
				return v, nil
			}

			if trustCenterDoc.OriginalFileID == nil {
				return nil, errMissingFileID
			}

			tuples := []fgax.TupleKey{}

			if trustCenterDoc.Visibility != enums.TrustCenterDocumentVisibilityNotVisible {
				/// If the document is "visible", add the wildcard viewer tuple for the document
				tuples = append(tuples, fgax.CreateWildcardViewerTuple(trustCenterDoc.ID, "trust_center_doc")...)

				if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
					// Files are only globally viewable if the document is publicly visible
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.OriginalFileID, generated.TypeFile)...)
				}
			}

			if len(tuples) > 0 {
				if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
					return nil, err
				}
			}

			if trustCenterDoc.WatermarkingEnabled {
				zerolog.Ctx(ctx).Debug().Msg("watermarking enabled, queuing job")
				if _, err := m.Job.Insert(ctx, corejobs.WatermarkDocArgs{
					TrustCenterDocumentID: trustCenterDoc.ID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return trustCenterDoc, nil
		})
	}, ent.OpCreate)
}

func HookUpdateTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			// Skip hook logic if this is an internal operation from the create hook
			if _, isInternal := contextx.From[internalTrustCenterDocUpdateKey](ctx); isInternal {
				zerolog.Ctx(ctx).Debug().Msg("skipping update hook for internal operation")
				return next.Mutate(ctx, m)
			}

			zerolog.Ctx(ctx).Debug().Msg("trust center doc hook")

			var err error
			var fileUploaded bool
			ctx, fileUploaded, err = checkTrustCenterDocFile(ctx, m)
			if err != nil {
				return nil, err
			}

			_, mutationSetsOriginalFileID := m.OriginalFileID()

			var watermarkFileUploaded bool
			ctx, watermarkFileUploaded, err = checkWatermarkedTrustCenterDocFile(ctx, m)
			if err != nil {
				return nil, err
			}

			_, mutationSetFileID := m.FileID()

			zerolog.Ctx(ctx).Debug().Bool("file_uploaded", fileUploaded).Bool("watermark_file_uploaded", watermarkFileUploaded).Bool("mutation_sets_original_file_id", mutationSetsOriginalFileID).Bool("mutation_set_file_id", mutationSetFileID).Msg("trust center doc hook")

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			trustCenterDoc, ok := v.(*generated.TrustCenterDoc)
			if !ok {
				return v, nil
			}

			if trustCenterDoc.OriginalFileID == nil {
				return nil, errMissingFileID
			}

			if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
				tuples := []fgax.TupleKey{}
				if (mutationSetFileID || watermarkFileUploaded) && *trustCenterDoc.FileID != *trustCenterDoc.OriginalFileID {
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.FileID, generated.TypeFile)...)
				}

				if mutationSetsOriginalFileID || fileUploaded {
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.OriginalFileID, generated.TypeFile)...)
				}

				if len(tuples) > 0 {
					if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
						return nil, err
					}
				}
			}

			fileIDsToUpdate := []string{*trustCenterDoc.OriginalFileID}
			if trustCenterDoc.FileID != nil && *trustCenterDoc.FileID != *trustCenterDoc.OriginalFileID {
				fileIDsToUpdate = append(fileIDsToUpdate, *trustCenterDoc.FileID)
			}

			if err = updateTrustCenterDocVisibility(ctx, m, fileIDsToUpdate, trustCenterDoc.ID); err != nil {
				return nil, err
			}

			if mutationSetsOriginalFileID || fileUploaded {
				if trustCenterDoc.WatermarkingEnabled {
					if _, err := m.Job.Insert(ctx, corejobs.WatermarkDocArgs{
						TrustCenterDocumentID: trustCenterDoc.ID,
					}, nil); err != nil {
						return nil, err
					}
				} else {
					// Use privacy allow context for internal update operation to bypass authorization checks
					// and mark as internal operation to avoid triggering the update hook logic
					allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
					internalCtx := contextx.With(allowCtx, internalTrustCenterDocUpdateKey{})
					trustCenterDoc, err = m.Client().TrustCenterDoc.UpdateOne(trustCenterDoc).SetFileID(*trustCenterDoc.OriginalFileID).Save(internalCtx)
					if err != nil {
						return nil, err
					}
				}
			}

			return trustCenterDoc, nil
		})
	}, ent.OpUpdateOne)
}

// updateTrustCenterDocVisibility updates fga tuples based on the visibility of the trust center doc
func updateTrustCenterDocVisibility(ctx context.Context, m *generated.TrustCenterDocMutation, fileIDs []string, docID string) error {
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
		for _, fileID := range fileIDs {
			writes = append(writes, fgax.CreateWildcardViewerTuple(fileID, generated.TypeFile)...)
		}
	}

	// 3. If the visibility has changed from public to not visible or protected, remove the wildcard tuples
	if oldVisibility == enums.TrustCenterDocumentVisibilityPubliclyVisible &&
		(visibility == enums.TrustCenterDocumentVisibilityNotVisible || visibility == enums.TrustCenterDocumentVisibilityProtected) {
		for _, fileID := range fileIDs {
			deletes = append(deletes, fgax.CreateWildcardViewerTuple(fileID, generated.TypeFile)...)
		}
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
func checkTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, bool, error) {
	dockey := "trustCenterDocFile"
	uploadsFile := false

	// get the file from the context, if it exists
	docFile, _ := objects.FilesFromContextWithKey(ctx, dockey)
	if docFile == nil {
		return ctx, uploadsFile, nil
	}

	// this should always be true, but check just in case
	if docFile[0].FieldName == dockey {
		// we should only have one file
		if len(docFile) > 1 {
			return ctx, uploadsFile, ErrNotSingularUpload
		}
		m.SetOriginalFileID(docFile[0].ID)

		docFile[0].Parent.ID, _ = m.ID()
		docFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, dockey, docFile[0])
		uploadsFile = true
	}

	return ctx, uploadsFile, nil
}

// checkWatermarkedTrustCenterDocFile checks if the doc file is uploaded, and sets the file ID
func checkWatermarkedTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, bool, error) {
	dockey := "watermarkedTrustCenterDocFile"
	uploadsFile := false

	// get the file from the context, if it exists
	docFile, _ := objects.FilesFromContextWithKey(ctx, dockey)
	if docFile == nil {
		return ctx, uploadsFile, nil
	}

	// this should always be true, but check just in case
	if docFile[0].FieldName == dockey {
		// we should only have one file
		if len(docFile) > 1 {
			return ctx, uploadsFile, ErrNotSingularUpload
		}
		m.SetFileID(docFile[0].ID)

		docFile[0].Parent.ID, _ = m.ID()
		docFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, dockey, docFile[0])
		uploadsFile = true
	}

	return ctx, uploadsFile, nil
}
