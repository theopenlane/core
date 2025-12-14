package hooks

import (
	"context"
	"errors"

	"entgo.io/ent"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

var (
	errMissingFileID         = errors.New("missing file id")
	errCannotSetFileOnCreate = errors.New("cannot set file id on create")
	errNotSingularUpload     = errors.New("expected a single file upload")
)

// internalTrustCenterDocUpdateKey is used to mark internal update operations within hooks
type internalTrustCenterDocUpdateKey struct{}

// HookCreateTrustCenterDoc is an ent hook that processes file uploads and sets appropriate fields and permissions on create
func HookCreateTrustCenterDoc() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center doc create hook")
			// check for uploaded files (e.g. logo image)
			fileIDs, _ := objects.FilesFromContextWithKey(ctx, "watermarkedTrustCenterDocFile")

			_, mutationSetsFileID := m.FileID()
			_, mutationSetsOriginalFileID := m.OriginalFileID()

			if mutationSetsFileID || len(fileIDs) > 0 {
				return nil, errCannotSetFileOnCreate
			}

			// Process trust center doc file
			docFiles, _ := objects.FilesFromContextWithKey(ctx, "trustCenterDocFile")
			if len(docFiles) > 0 {
				var err error
				ctx, err = checkTrustCenterDocFile(ctx, m)
				if err != nil {
					return nil, err
				}

				// Get files again after processing to get updated parent info
				docFiles, _ = objects.FilesFromContextWithKey(ctx, "trustCenterDocFile")

				// we should only have one file
				if len(docFiles) > 1 {
					return nil, errNotSingularUpload
				}

				m.SetOriginalFileID(docFiles[0].ID)
			}

			_, watermarkingEnabledSet := m.WatermarkingEnabled()
			trustCenterID, trustCenterIDSet := m.TrustCenterID()

			if trustCenterIDSet {
				// check the config to see if watermarking is enabled by default
				config, err := m.Client().TrustCenterWatermarkConfig.Query().
					Where(trustcenterwatermarkconfig.TrustCenterID(trustCenterID)).
					Only(ctx)
				if err == nil && config != nil {
					// If global config is enabled, force document-level watermarking to true
					if config.IsEnabled {
						m.SetWatermarkingEnabled(true)
					} else if !watermarkingEnabledSet {
						// Global config is disabled, and user didn't explicitly set value, use config default
						m.SetWatermarkingEnabled(config.IsEnabled)
					}
				}
			}

			// Get the final watermarking enabled value after config overrides
			finalWatermarkingEnabled, _ := m.WatermarkingEnabled()

			// if we have no uploaded files
			if !mutationSetsOriginalFileID && len(docFiles) == 0 {
				// check if watermarking is enabled because if it is a file must be present
				if finalWatermarkingEnabled {
					return nil, errMissingFileID
				}

				// otherwise set the visibility to NOT_VISIBLE
				m.SetVisibility(enums.TrustCenterDocumentVisibilityNotVisible)
			} else if !finalWatermarkingEnabled {
				// Watermarking is disabled, so set file_id = original_file_id
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

			tuples := []fgax.TupleKey{}

			if trustCenterDoc.Visibility != enums.TrustCenterDocumentVisibilityNotVisible {
				/// If the document is "visible", add the wildcard viewer tuple for the document
				tuples = append(tuples, fgax.CreateWildcardViewerTuple(trustCenterDoc.ID, "trust_center_doc")...)

				if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible && trustCenterDoc.OriginalFileID != nil {
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
				logx.FromContext(ctx).Debug().Msg("watermarking enabled, queuing job")
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

// HookUpdateTrustCenterDoc is an ent hook that processes file uploads and sets appropriate fields and permissions on update
func HookUpdateTrustCenterDoc() ent.Hook { // nolint:gocyclo
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterDocFunc(func(ctx context.Context, m *generated.TrustCenterDocMutation) (generated.Value, error) {
			// Skip hook logic if this is an internal operation from the create hook
			if _, isInternal := contextx.From[internalTrustCenterDocUpdateKey](ctx); isInternal {
				logx.FromContext(ctx).Debug().Msg("skipping update hook for internal operation")
				return next.Mutate(ctx, m)
			}

			logx.FromContext(ctx).Debug().Msg("trust center doc hook")

			// Check if watermarking is being modified and enforce global config override rules
			if newWatermarkingEnabled, watermarkingEnabledSet := m.WatermarkingEnabled(); watermarkingEnabledSet {
				// Get the trust center ID - for updates we need to look it up from the document
				docID, docIDSet := m.ID()
				if docIDSet {
					doc, err := m.Client().TrustCenterDoc.Get(ctx, docID)
					if err == nil && doc != nil {
						config, err := m.Client().TrustCenterWatermarkConfig.Query().Where(trustcenterwatermarkconfig.TrustCenterID(doc.TrustCenterID)).Only(ctx)
						if err == nil && config != nil {
							// If global config is enabled, force document-level watermarking to true
							if config.IsEnabled && !newWatermarkingEnabled {
								logx.FromContext(ctx).Debug().Msg("global watermark config is enabled, forcing document watermarking to true")
								m.SetWatermarkingEnabled(true)
							}
						}
					}
				}
			}

			// Process trust center doc file
			docFiles, _ := objects.FilesFromContextWithKey(ctx, "trustCenterDocFile")
			if len(docFiles) > 0 {
				var err error
				ctx, err = checkTrustCenterDocFile(ctx, m)
				if err != nil {
					return nil, err
				}

				// Get files again after processing to get updated parent info
				docFiles, _ = objects.FilesFromContextWithKey(ctx, "trustCenterDocFile")

				// we should only have one file
				if len(docFiles) > 1 {
					return nil, ErrNotSingularUpload
				}

				m.SetOriginalFileID(docFiles[0].ID)
			}

			_, mutationSetsOriginalFileID := m.OriginalFileID()

			// Process watermarked file
			watermarkedFiles, _ := objects.FilesFromContextWithKey(ctx, "watermarkedTrustCenterDocFile")
			if len(watermarkedFiles) > 0 {
				var err error
				ctx, err = checkWatermarkedTrustCenterDocFile(ctx, m)
				if err != nil {
					return nil, err
				}

				// Get files again after processing to get updated parent info
				watermarkedFiles, _ = objects.FilesFromContextWithKey(ctx, "watermarkedTrustCenterDocFile")

				// we should only have one file
				if len(watermarkedFiles) > 1 {
					return nil, ErrNotSingularUpload
				}

				m.SetFileID(watermarkedFiles[0].ID)
			}

			_, mutationSetFileID := m.FileID()

			logx.FromContext(ctx).Debug().Bool("file_uploaded", len(docFiles) > 0).Bool("watermark_file_uploaded", len(watermarkedFiles) > 0).Bool("mutation_sets_original_file_id", mutationSetsOriginalFileID).Bool("mutation_set_file_id", mutationSetFileID).Msg("trust center doc hook")

			isExistingFileCleared := m.OriginalFileIDCleared()

			// if original file was cleared, automatically set visibility to NOT_VISIBLE
			if isExistingFileCleared {
				m.SetVisibility(enums.TrustCenterDocumentVisibilityNotVisible)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}

			trustCenterDoc, ok := v.(*generated.TrustCenterDoc)
			if !ok {
				return v, nil
			}

			if trustCenterDoc.Visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible {
				tuples := []fgax.TupleKey{}
				if trustCenterDoc.FileID != nil && trustCenterDoc.OriginalFileID != nil {
					if (mutationSetFileID || len(watermarkedFiles) > 0) && *trustCenterDoc.FileID != *trustCenterDoc.OriginalFileID {
						tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.FileID, generated.TypeFile)...)
					}
				}

				if trustCenterDoc.OriginalFileID != nil && (mutationSetsOriginalFileID || len(docFiles) > 0) {
					tuples = append(tuples, fgax.CreateWildcardViewerTuple(*trustCenterDoc.OriginalFileID, generated.TypeFile)...)
				}

				if len(tuples) > 0 {
					if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
						return nil, err
					}
				}
			}

			var fileIDsToUpdate []string

			if trustCenterDoc.OriginalFileID != nil {
				fileIDsToUpdate = []string{*trustCenterDoc.OriginalFileID}

				if trustCenterDoc.FileID != nil && *trustCenterDoc.FileID != *trustCenterDoc.OriginalFileID {
					fileIDsToUpdate = append(fileIDsToUpdate, *trustCenterDoc.FileID)
				}
			}

			// only update visibility tuples if it was explicitly changed
			if !isExistingFileCleared {
				if err = updateTrustCenterDocVisibility(ctx, m, fileIDsToUpdate, trustCenterDoc.ID); err != nil {
					return nil, err
				}
			}

			if mutationSetsOriginalFileID || len(docFiles) > 0 {
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
			logx.FromContext(ctx).Error().Err(err).Msg("failed to update file access permissions")
			return err
		}
	}
	return nil
}

// checkTrustCenterDocFile checks if a trust center doc file is provided and sets the parent information
func checkTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, error) {
	key := "trustCenterDocFile"

	// Create adapter for the existing mutation interface
	adapter := objects.NewGenericMutationAdapter(m,
		func(mut *generated.TrustCenterDocMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustCenterDocMutation) string { return mut.Type() },
	)

	// Use the generic helper to process files
	return objects.ProcessFilesForMutation(ctx, adapter, key, "trust_center_doc")
}

// checkWatermarkedTrustCenterDocFile checks if the watermarked doc file is provided and sets the parent information
func checkWatermarkedTrustCenterDocFile(ctx context.Context, m *generated.TrustCenterDocMutation) (context.Context, error) {
	key := "watermarkedTrustCenterDocFile"

	// Create adapter for the existing mutation interface
	adapter := objects.NewGenericMutationAdapter(m,
		func(mut *generated.TrustCenterDocMutation) (string, bool) { return mut.ID() },
		func(mut *generated.TrustCenterDocMutation) string { return mut.Type() },
	)

	// Use the generic helper to process files
	return objects.ProcessFilesForMutation(ctx, adapter, key, "trust_center_doc")
}
