package observability

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// emitTyped wraps and emits a typed event, recording emit errors using the provided operation
func emitTyped[T any](ctx context.Context, observer *Observer, emitter soiree.Emitter, topic soiree.TypedTopic[T], payload T, client any, op Operation, fields Fields) error {
	event, err := topic.Wrap(payload)
	if err != nil {
		return err
	}

	event.SetContext(ctx)
	if client != nil {
		event.SetClient(client)
	}

	fields = lo.Assign(Fields{FieldPayload: payload}, fields)
	errCh := emitter.Emit(topic.Name(), event)

	return observer.handleEmit(ctx, op, fields, topic.Name(), errCh)
}

// emitTypedFromScope wraps and emits a typed event, recording emit errors against the scope
func emitTypedFromScope[T any](scope *Scope, emitter soiree.Emitter, topic soiree.TypedTopic[T], payload T, client any, fields Fields) error {
	return emitTyped(scope.ctx, scope.observer, emitter, topic, payload, client, scope.op, fields)
}

// EmitFromTopic wraps and emits a typed event using topic-provided fields without initializing a scope
func EmitFromTopic[T any](scope *Scope, emitter soiree.Emitter, topic soiree.TypedTopic[T], payload T, client any, extra Fields) error {
	return emitTypedFromScope(scope, emitter, topic, payload, client, extra)
}

// emitWithPayload wraps and emits a typed event with standard operation metadata
func emitWithPayload[T any](ctx context.Context, observer *Observer, emitter soiree.Emitter, topic soiree.TypedTopic[T], payload T, client any, op OperationName, origin Origin, triggerEvent string, fields Fields) error {
	return emitTyped(ctx, observer, emitter, topic, payload, client, Operation{
		Name:         op,
		Origin:       origin,
		TriggerEvent: triggerEvent,
	}, fields)
}

// EmitEngine wraps and emits a typed event for engine operations
func EmitEngine[T any](ctx context.Context, observer *Observer, emitter soiree.Emitter, topic soiree.TypedTopic[T], payload T, client any, op OperationName, triggerEvent string, fields Fields) error {
	return emitWithPayload(ctx, observer, emitter, topic, payload, client, op, OriginEngine, triggerEvent, fields)
}
