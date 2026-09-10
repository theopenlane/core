package hooks

import (
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// TrustCenterWatermarkListeners enqueues watermarking jobs when trust center document files change
func TrustCenterWatermarkListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaTrustCenterDoc,
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Fields:     []string{trustcenterdoc.FieldOriginalFileID},
			Handle:     handleTrustCenterDocWatermarkGala,
		},
	}
}

// handleTrustCenterDocWatermarkGala enqueues a watermarking job for a trust center document if watermarking is enabled
func handleTrustCenterDocWatermarkGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	document, err := inv.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.ID(inv.EntityID)).
		Where(trustcenterdoc.WatermarkingEnabled(true)).
		Select(trustcenterdoc.FieldID).
		Only(inv.Context)
	if err != nil {
		if generated.IsNotFound(err) {
			logx.FromContext(inv.Context).Info().Msg("trust center document not found or watermarking disabled, skipping watermark job")

			return nil
		}

		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to query trust center document for watermark job")

		return err
	}

	logx.FromContext(inv.Context).Debug().Msg("watermarking enabled, queuing job")

	if err := enqueueJob(inv.Context, inv.Client.Job, jobspec.WatermarkDocArgs{
		TrustCenterDocumentID: document.ID,
	}, nil); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed to enqueue trust center doc watermark job")

		return err
	}

	return nil
}
