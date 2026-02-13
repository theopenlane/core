package hooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gala"
)

// enqueueGalaMutationOutbox builds and dispatches a durable gala envelope for a mutation event.
func enqueueGalaMutationOutbox(
	ctx context.Context,
	runtime *gala.Runtime,
	topic string,
	payload *events.MutationPayload,
	props soiree.Properties,
) error {
	if runtime == nil {
		return ErrGalaRuntimeUnavailable
	}

	mutationTopic := gala.Topic[eventqueue.MutationGalaPayload]{
		Name: gala.TopicName(topic),
	}

	if err := ensureGalaMutationTopicRegistered(runtime.Registry(), mutationTopic); err != nil {
		return fmt.Errorf("%w: %w", ErrGalaMutationOutboxEnqueueFailed, err)
	}

	envelope, err := eventqueue.NewMutationGalaEnvelope(context.WithoutCancel(ctx), runtime, mutationTopic, payload, props)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGalaMutationOutboxEnqueueFailed, err)
	}

	if err := runtime.EmitEnvelope(context.WithoutCancel(ctx), envelope); err != nil {
		return fmt.Errorf("%w: %w", ErrGalaMutationOutboxEnqueueFailed, err)
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
