package system

import (
	"context"
	"time"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/objects/filerefs"
)

// Run hard deletes unreferenced file records older than FileOrphanMinAge, soft deleted ones
// included, or logs them under dry run, and returns the count
func (s FileOrphanSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	if s.MaxPerRun <= 0 {
		s.MaxPerRun = DefaultFileOrphanMaxPerRun
	}

	systemCtx := entx.SkipSoftDelete(auth.WithSystemSweepContext(ctx))
	cutoff := time.Now().Add(-FileOrphanMinAge)

	var (
		processed int
		last      string
	)

	for processed < s.MaxPerRun {
		ids, err := req.DB.File.Query().
			Where(file.CreatedAtLTE(cutoff), file.IDGT(last)).
			Order(ent.Asc(file.FieldID)).
			Limit(fileOrphanPageSize).
			IDs(systemCtx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed querying files for orphan sweep")

			return processed, err
		}

		if len(ids) == 0 {
			break
		}

		last = ids[len(ids)-1]

		// soft deleted parents still own their files
		referenced, err := filerefs.Referenced(systemCtx, req.DB, ids)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed classifying files for orphan sweep")

			return processed, err
		}

		for _, id := range ids {
			if _, ok := referenced[id]; ok {
				continue
			}

			if processed >= s.MaxPerRun {
				break
			}

			if s.DryRun {
				logx.FromContext(ctx).Info().Str("file_id", id).Msg("dry run: would delete orphaned file")
				processed++

				continue
			}

			if err := req.DB.File.DeleteOneID(id).Exec(systemCtx); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("file_id", id).Msg("failed to delete orphaned file")

				continue
			}

			logx.FromContext(ctx).Info().Str("file_id", id).Msg("deleted orphaned file")
			processed++
		}
	}

	logx.FromContext(ctx).Info().Int("count", processed).Bool("dry_run", s.DryRun).Msg("file orphan sweep summary")

	return processed, nil
}
