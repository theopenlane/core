package observability

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// EmitTyped emits a typed payload through Gala, recording emit errors with operation metadata.
func EmitTyped[T any](ctx context.Context, observer *Observer, runtime *gala.Gala, topic gala.TopicName, payload T, op Operation, fields Fields) error {
	return emitTyped(ctx, observer, runtime, topic, payload, op, fields)
}

// emitTyped emits a typed payload through Gala, recording emit errors with operation metadata
func emitTyped[T any](ctx context.Context, observer *Observer, runtime *gala.Gala, topic gala.TopicName, payload T, op Operation, fields Fields) error {
	if runtime == nil {
		return nil
	}

	fields = lo.Assign(Fields{FieldPayload: payload}, fields)

	if _, err := runtime.EmitWithHeaders(ctx, topic, payload, gala.Headers{}); err != nil {
		observer.handleEmitError(ctx, op, fields, string(topic), err)

		return err
	}

	return nil
}
