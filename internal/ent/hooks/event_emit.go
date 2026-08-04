package hooks

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/pkg/gala"
)

// enqueueGalaMutation dispatches a durable gala envelope for a mutation event.
// The mutation event id becomes the envelope id so durable dispatch is traceable
// per mutation; topic registration is guaranteed by the InterestedIn gate in the
// emit hook (listener registration auto-registers the topic)
func enqueueGalaMutation(ctx context.Context, g *gala.Gala, topic string, payload eventqueue.MutationGalaPayload, metadata eventqueue.MutationGalaMetadata) error {
	if g == nil {
		return ErrGalaRuntimeUnavailable
	}

	// detach cancellation for best-effort dispatch after commit
	dispatchCtx := context.WithoutCancel(ctx)

	if _, err := g.Emit(dispatchCtx, gala.TopicName(topic), payload,
		gala.WithHeaders(eventqueue.NewGalaHeadersFromMutationMetadata(metadata)),
		gala.WithEventID(gala.EventID(metadata.EventID))); err != nil {
		return fmt.Errorf("%w: emit: %v", ErrGalaMutationEnqueueFailed, err)
	}

	return nil
}
