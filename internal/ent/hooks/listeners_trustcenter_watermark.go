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
func RegisterGalaTrustCenterWatermarkListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema:     generated.TypeTrustCenterDoc,
			Name:       "trustcenter.watermark.doc",
			Operations: []string{ent.OpCreate.String(), ent.OpUpdateOne.String()},
			Handle:     handleTrustCenterDocWatermarkGala,
		},
	)
}

func handleTrustCenterDocWatermarkGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	if payload.Operation == ent.OpUpdateOne.String() && !eventqueue.MutationFieldChanged(payload, trustcenterdoc.FieldOriginalFileID) {
		logx.FromContext(inv.Context).Info().Msg("no original file change detected in mutation, skipping trust center doc watermark job")
		return nil
	}

	document, err := inv.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.ID(inv.EntityID)).
		Where(trustcenterdoc.WatermarkingEnabled(true)).
		Select(trustcenterdoc.FieldID).
		Only(inv.Context)
	if err != nil {
		if generated.IsNotFound(err) {
			logx.FromContext(inv.Context).Info().
				Str("trust_center_doc_id", inv.EntityID).
				Msg("trust center document not found or watermarking disabled, skipping watermark job")
			return nil
		}

		logx.FromContext(inv.Context).Error().
			Err(err).
			Str("trust_center_doc_id", inv.EntityID).
			Msg("failed to query trust center document for watermark job")

		return err
	}

	logx.FromContext(inv.Context).Debug().
		Str("trust_center_doc_id", document.ID).
		Msg("watermarking enabled, queuing job")

	if err := enqueueJob(inv.Context, inv.Client.Job, jobspec.WatermarkDocArgs{
		TrustCenterDocumentID: document.ID,
	}, nil); err != nil {
		logx.FromContext(inv.Context).Error().
			Err(err).
			Str("trust_center_doc_id", document.ID).
			Msg("failed to enqueue trust center doc watermark job")

		return err
	}

	return nil
}
