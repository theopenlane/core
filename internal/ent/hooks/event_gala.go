package hooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/gala"
)

// enqueueGalaMutation builds and dispatches a durable gala envelope for a mutation event.
func enqueueGalaMutation(
	ctx context.Context,
	runtime *gala.Runtime,
	topic string,
	payload *events.MutationPayload,
	metadata eventqueue.MutationGalaMetadata,
) error {
	if runtime == nil {
		return ErrGalaRuntimeUnavailable
	}

	mutationTopic := gala.Topic[eventqueue.MutationGalaPayload]{
		Name: gala.TopicName(topic),
	}

	if err := ensureGalaMutationTopicRegistered(runtime.Registry(), mutationTopic); err != nil {
		return fmt.Errorf("%w: %w", ErrGalaMutationEnqueueFailed, err)
	}

	envelope, err := eventqueue.NewMutationGalaEnvelope(context.WithoutCancel(ctx), runtime, mutationTopic, payload, metadata)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGalaMutationEnqueueFailed, err)
	}

	if err := runtime.EmitEnvelope(context.WithoutCancel(ctx), envelope); err != nil {
		return fmt.Errorf("%w: %w", ErrGalaMutationEnqueueFailed, err)
	}

	return nil
}

// ensureGalaMutationTopicRegistered ensures a mutation topic contract is available in the gala registry.
func ensureGalaMutationTopicRegistered(registry *gala.Registry, topic gala.Topic[eventqueue.MutationGalaPayload]) error {
	registration := gala.Registration[eventqueue.MutationGalaPayload]{
		Topic: topic,
		Codec: gala.JSONCodec[eventqueue.MutationGalaPayload]{},
		Policy: gala.TopicPolicy{
			EmitMode:   gala.EmitModeDurable,
			QueueClass: gala.QueueClassWorkflow,
		},
	}

	err := registration.Register(registry)
	if err == nil {
		return nil
	}

	if errors.Is(err, gala.ErrTopicAlreadyRegistered) {
		return nil
	}

	return err
}
