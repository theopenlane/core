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
		return nil
	}

	if payload.Operation == ent.OpUpdateOne.String() && !eventqueue.MutationFieldChanged(payload, trustcenterdoc.FieldOriginalFileID) {
		return nil
	}

	documentID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || documentID == "" {
		return nil
	}

	document, err := client.TrustCenterDoc.Query().
		Where(trustcenterdoc.ID(documentID)).
		Select(trustcenterdoc.FieldID, trustcenterdoc.FieldWatermarkingEnabled).
		Only(ctx.Context)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	if !document.WatermarkingEnabled {
		return nil
	}

	logx.FromContext(ctx.Context).Debug().
		Str("trust_center_doc_id", document.ID).
		Msg("watermarking enabled, queuing job")

	return enqueueJob(ctx.Context, client.Job, jobspec.WatermarkDocArgs{
		TrustCenterDocumentID: document.ID,
	}, nil)
}
