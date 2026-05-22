package hooks

import (
	"entgo.io/ent"

	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaTrustCenterWatermarkListeners registers listeners that enqueue
// watermarking jobs after trust center document db transactions have been committed.
func RegisterGalaTrustCenterWatermarkListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeTrustCenterDoc),
			Name:       "trustcenter.watermark.doc",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdateOne.String()},
			Handle:     handleTrustCenterDocWatermarkGala,
		},
	)
}

func handleTrustCenterDocWatermarkGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		logx.FromContext(ctx.Context).Info().Msg("no ent client found in gala handler context, skipping trust center doc watermark job")
		return nil
	}

	if payload.Operation == ent.OpUpdateOne.String() && !eventqueue.MutationFieldChanged(payload, trustcenterdoc.FieldOriginalFileID) {
		logx.FromContext(ctx.Context).Info().Msg("no original file change detected in mutation, skipping trust center doc watermark job")
		return nil
	}

	documentID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || documentID == "" {
		logx.FromContext(ctx.Context).Info().Msg("no trust center document ID found in mutation, skipping watermark job")
		return nil
	}

	document, err := client.TrustCenterDoc.Query().
		Where(trustcenterdoc.ID(documentID)).
		Where(trustcenterdoc.WatermarkingEnabled(true)).
		Select(trustcenterdoc.FieldID).
		Only(ctx.Context)
	if err != nil {
		if generated.IsNotFound(err) {
			logx.FromContext(ctx.Context).Info().
				Str("trust_center_doc_id", documentID).
				Msg("trust center document not found or watermarking disabled, skipping watermark job")
			return nil
		}

		logx.FromContext(ctx.Context).Error().
			Err(err).
			Str("trust_center_doc_id", documentID).
			Msg("failed to query trust center document for watermark job")

		return err
	}

	logx.FromContext(ctx.Context).Debug().
		Str("trust_center_doc_id", document.ID).
		Msg("watermarking enabled, queuing job")

	if err := enqueueJob(ctx.Context, client.Job, jobspec.WatermarkDocArgs{
		TrustCenterDocumentID: document.ID,
	}, nil); err != nil {
		logx.FromContext(ctx.Context).Error().
			Err(err).
			Str("trust_center_doc_id", document.ID).
			Msg("failed to enqueue trust center doc watermark job")

		return err
	}

	return nil
}
