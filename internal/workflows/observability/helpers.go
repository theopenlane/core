package observability

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/core/pkg/gala"
)

// BeginListenerTopic starts an observation for a workflow listener using topic metadata
func BeginListenerTopic[T any](ctx gala.HandlerContext, observer *Observer, topic gala.TopicName, payload T, extra Fields) *Scope {
	baseCtx := context.Background()
	if ctx.Context != nil {
		baseCtx = ctx.Context
	}

	opName := OperationName(topic)
	triggerEvent := string(ctx.Envelope.Topic)
	fields := lo.Assign(Fields{FieldPayload: payload}, extra)

	return observer.begin(baseCtx, Operation{
		Name:         opName,
		Origin:       OriginListeners,
		TriggerEvent: triggerEvent,
	}, fields)
}

// BeginEngine starts an observation for workflow engine operations
func BeginEngine(ctx context.Context, observer *Observer, op OperationName, triggerEvent string, fields Fields) *Scope {
	return observer.begin(ctx, Operation{
		Name:         op,
		Origin:       OriginEngine,
		TriggerEvent: triggerEvent,
	}, fields)
}
