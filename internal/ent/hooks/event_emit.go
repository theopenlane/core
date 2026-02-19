package hooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/pkg/gala"
)

// enqueueGalaMutation builds and dispatches a durable gala envelope for a mutation event
func enqueueGalaMutation(ctx context.Context, g *gala.Gala, topic string, payload eventqueue.MutationGalaPayload, metadata eventqueue.MutationGalaMetadata) error {
	if g == nil {
		return ErrGalaRuntimeUnavailable
	}

	mutationTopic := gala.Topic[eventqueue.MutationGalaPayload]{
		Name: gala.TopicName(topic),
	}

	if err := ensureGalaMutationTopicRegistered(g.Registry(), mutationTopic); err != nil {
		return fmt.Errorf("%w: topic registration: %v", ErrGalaMutationEnqueueFailed, err)
	}

	// detach cancellation for best-effort dispatch after commit
	dispatchCtx := context.WithoutCancel(ctx)

	envelope, err := eventqueue.NewMutationGalaEnvelope(dispatchCtx, g, mutationTopic, payload, metadata)
	if err != nil {
		return fmt.Errorf("%w: envelope construction: %v", ErrGalaMutationEnqueueFailed, err)
	}

	if err := g.EmitEnvelope(dispatchCtx, envelope); err != nil {
		return fmt.Errorf("%w: emit: %v", ErrGalaMutationEnqueueFailed, err)
	}

	return nil
}

// ensureGalaMutationTopicRegistered ensures a mutation topic contract is available in the gala registry.
func ensureGalaMutationTopicRegistered(registry *gala.Registry, topic gala.Topic[eventqueue.MutationGalaPayload]) error {
	err := gala.RegisterTopic(registry, gala.Registration[eventqueue.MutationGalaPayload]{
		Topic: topic,
		Codec: gala.JSONCodec[eventqueue.MutationGalaPayload]{},
	})

	if errors.Is(err, gala.ErrTopicAlreadyRegistered) {
		return nil
	}

	return err
}
