package hooks

import (
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// TrustCenterWatermarkListeners returns the listeners that enqueue watermarking jobs
// after trust center document transactions commit
func TrustCenterWatermarkListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaTrustCenterDoc,
			Label:      "watermark",
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Fields:     []string{trustcenterdoc.FieldOriginalFileID},
			Handle:     handleTrustCenterDocWatermarkGala,
		},
	}
}

// handleTrustCenterDocWatermarkGala enqueues a watermarking job for a trust center document if watermarking is enabled
func handleTrustCenterDocWatermarkGala(inv entityops.Invocation, payload entityops.MutationPayload) error {
	ctx := inv.Context

	document, err := inv.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.ID(inv.EntityID)).
		Where(trustcenterdoc.WatermarkingEnabled(true)).
		Select(trustcenterdoc.FieldID).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			logx.FromContext(ctx).Info().Msg("trust center document not found or watermarking disabled, skipping watermark job")
			return nil
		}

		logx.FromContext(ctx).Error().Err(err).Msg("failed to query trust center document for watermark job")

		return err
	}

	logx.FromContext(ctx).Debug().Msg("watermarking enabled, queuing job")

	if err := enqueueJob(ctx, inv.Client.Job, jobspec.WatermarkDocArgs{
		TrustCenterDocumentID: document.ID,
	}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to enqueue trust center doc watermark job")

		return err
	}

	return nil
}
