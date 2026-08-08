package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaTrustCenterWatermarkListeners registers listeners that enqueue
// watermarking jobs after trust center document db transactions have been committed.
func RegisterGalaTrustCenterWatermarkListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g,
		entityops.MutationListener{
			Schema:     generated.TypeTrustCenterDoc,
			Operations: []string{ent.OpCreate.String(), ent.OpUpdateOne.String()},
			Fields:     []string{trustcenterdoc.FieldOriginalFileID},
			Enrich: func(ctx context.Context, payload entityops.MutationPayload) context.Context {
				return logx.WithFields(ctx, map[string]any{"trust_center_doc_id": payload.EntityID})
			},
			Handle: handleTrustCenterDocWatermarkGala,
		},
	)
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
