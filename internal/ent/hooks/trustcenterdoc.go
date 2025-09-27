package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/rs/zerolog"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/iam/fgax"
)

var (
	errMissingFileID                     = errors.New("missing file id")
	errCannotUpdateWatermarkFileOnCreate = errors.New("cannot update watermark file on create")
)

func HookUpdateTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("update trust center doc hook")

			ctx, updatedOriginalFile, err := checkTrustCenterDocOriginalFile(ctx, m)
			if err != nil {
				return nil, err
			}

			ctx, updatedWatermarkedFile, err := checkTrustCenterDocWatermarkedFile(ctx, m)
			if err != nil {
				return nil, err
			}

			if updatedWatermarkedFile {
				zerolog.Ctx(ctx).Debug().Msg("watermarked file updated on create")
			}
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			trustCenterDoc, ok := v.(*generated.TrustCenterDoc)
			if !ok {
				return v, nil
			}

			if updatedOriginalFile && !trustCenterDoc.WatermarkingEnabled {
				if err := m.Client().TrustCenterDoc.UpdateOne(trustCenterDoc).
					SetFileID(*trustCenterDoc.OriginalFileID).
					Exec(ctx); err != nil {
					return nil, err
				}
			} else if updatedOriginalFile && trustCenterDoc.WatermarkingEnabled {
				_, err := m.Job.Insert(ctx, corejobs.WatermarkDocArgs{
					TrustCenterDocumentID: trustCenterDoc.ID,
				}, nil)
				if err != nil {
					return nil, err
				}

			}

			if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
				tuples := []fgax.TupleKey{}

				if updatedOriginalFile {
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.OriginalFileID, generated.TypeFile)...)
				}

				if updatedWatermarkedFile {
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.FileID, generated.TypeFile)...)
				}

				if len(tuples) > 0 {
					if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
						return nil, err
					}
				}
			}

			if err := updateTrustCenterDocVisibility(ctx, m, []string{*trustCenterDoc.OriginalFileID}, trustCenterDoc.ID); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpUpdateOne)
}
func HookCreateTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			zerolog.Ctx(ctx).Debug().Msg("create trust center doc hook")

			// only allowed to upload a watermarked file on update
			wmFileID, mutationSetsFileID := m.FileID()
			wmStatus, mutationSetsWatermarkedStatus := m.WatermarkStatus()

			zerolog.Ctx(ctx).Debug().Msgf("wmFileID: %s, mutationSetsFileID: %t, wmStatus: %s, mutationSetsWatermarkedStatus: %t", wmFileID, mutationSetsFileID, wmStatus, mutationSetsWatermarkedStatus)
			if mutationSetsFileID || (mutationSetsWatermarkedStatus && wmStatus != enums.WatermarkStatusDisabled) {
				return nil, errCannotUpdateWatermarkFileOnCreate
			}

			// check for uploaded files (e.g. logo image)
			fileIDs := objects.GetFileIDsFromContext(ctx)

			_, mutationSetsOriginalFileID := m.OriginalFileID()

			// check if the operation is a create operation
			if len(fileIDs) < 1 && !mutationSetsOriginalFileID {
				return nil, objects.ErrNoFilesUploaded
			}

			if len(fileIDs) > 0 {
				var err error
				ctx, _, err = checkTrustCenterDocOriginalFile(ctx, m)
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

			if trustCenterDoc.OriginalFileID == nil {
				return nil, errMissingFileID
			}

			if trustCenterDoc.Visibility != enums.TrustCenterDocumentVisibilityNotVisible {
				tuples := fgax.CreateWildcardViewerTuple(trustCenterDoc.ID, "trust_center_doc")

				if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.OriginalFileID, generated.TypeFile)...)
				}

				if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
					return nil, err
				}
			}

			// If watermarking is enabled, kick off the watermarking job
			if trustCenterDoc.WatermarkingEnabled {
				if err := m.Client().TrustCenterDoc.UpdateOne(trustCenterDoc).
					SetWatermarkStatus(enums.WatermarkStatusPending).
					Exec(ctx); err != nil {
					return nil, err
				}
				m.Job.Insert(
					ctx,
					corejobs.WatermarkDocArgs{
						TrustCenterDocumentID: trustCenterDoc.ID,
					},
					nil,
				)
			} else {
				// otherwise, set the file ID to the original file ID
				if err := m.Client().TrustCenterDoc.UpdateOne(trustCenterDoc).
					SetFileID(*trustCenterDoc.OriginalFileID).
					Exec(ctx); err != nil {
					return nil, err
				}
			}

			return trustCenterDoc, nil
		})
	}, ent.OpCreate)
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
func checkTrustCenterDocOriginalFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, bool, error) {
	dockey := "trustCenterDocFile"

	// get the file from the context, if it exists
	originalDocFile, _ := objects.FilesFromContextWithKey(ctx, dockey)
	if originalDocFile == nil {
		return ctx, false, nil
	}
	updated := false

	// this should always be true, but check just in case
	if originalDocFile[0].FieldName == dockey {
		// we should only have one file
		if len(originalDocFile) > 1 {
			return ctx, false, ErrNotSingularUpload
		}
		m.SetOriginalFileID(originalDocFile[0].ID)

		originalDocFile[0].Parent.ID, _ = m.ID()
		originalDocFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, dockey, originalDocFile[0])
		updated = true
	}

	return ctx, updated, nil
}

// checkTrustCenterDocFile checks if a trust center doc file is provided and sets the local file ID
func checkTrustCenterDocWatermarkedFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, bool, error) {
	updated := false
	watermarkedDocKey := "watermarkedTrustCenterDocFile"
	watermarkedDocFile, _ := objects.FilesFromContextWithKey(ctx, watermarkedDocKey)
	if watermarkedDocFile != nil && watermarkedDocFile[0].FieldName == watermarkedDocKey {
		if len(watermarkedDocFile) > 1 {
			return ctx, updated, ErrNotSingularUpload
		}
		m.SetFileID(watermarkedDocFile[0].ID)

		watermarkedDocFile[0].Parent.ID, _ = m.ID()
		watermarkedDocFile[0].Parent.Type = "trust_center_doc"

		ctx = objects.UpdateFileInContextByKey(ctx, watermarkedDocKey, watermarkedDocFile[0])
		updated = true
	}

	return ctx, updated, nil
}
