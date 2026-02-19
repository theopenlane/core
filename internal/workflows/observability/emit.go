package observability

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/core/pkg/gala"
)

// emitTyped emits a typed payload through Gala, recording emit errors with operation metadata
func emitTyped[T any](ctx context.Context, observer *Observer, runtime *gala.Gala, topic gala.TopicName, payload T, op Operation, fields Fields) error {
	if runtime == nil {
		return nil
	}

	fields = lo.Assign(Fields{FieldPayload: payload}, fields)
	receipt := runtime.EmitWithHeaders(ctx, topic, payload, gala.Headers{})

	errCh := make(chan error, 1)
	if receipt.Err != nil {
		errCh <- receipt.Err
	}

	close(errCh)

	return observer.handleEmit(ctx, op, fields, string(topic), errCh)
}

// emitTypedFromScope emits a typed payload, recording emit errors against the supplied scope
func emitTypedFromScope[T any](scope *Scope, runtime *gala.Gala, topic gala.TopicName, payload T, fields Fields) error {
	return emitTyped(scope.ctx, scope.observer, runtime, topic, payload, scope.op, fields)
}

// EmitFromTopic emits a payload using topic metadata without creating a new scope
func EmitFromTopic[T any](scope *Scope, runtime *gala.Gala, topic gala.TopicName, payload T, extra Fields) error {
	return emitTypedFromScope(scope, runtime, topic, payload, extra)
}

// emitWithPayload emits a typed payload with explicit operation metadata
func emitWithPayload[T any](ctx context.Context, observer *Observer, runtime *gala.Gala, topic gala.TopicName, payload T, op OperationName, origin Origin, triggerEvent string, fields Fields) error {
	return emitTyped(ctx, observer, runtime, topic, payload, Operation{
		Name:         op,
		Origin:       origin,
		TriggerEvent: triggerEvent,
	}, fields)
}

// EmitEngine emits a typed payload for engine-originated operations
func EmitEngine[T any](ctx context.Context, observer *Observer, runtime *gala.Gala, topic gala.TopicName, payload T, op OperationName, triggerEvent string, fields Fields) error {
	return emitWithPayload(ctx, observer, runtime, topic, payload, op, OriginEngine, triggerEvent, fields)
}
