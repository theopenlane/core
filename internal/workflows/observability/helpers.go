package observability

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// BeginListenerTopic starts an observation for a workflow listener using topic metadata
func BeginListenerTopic[T any](observer *Observer, ctx *soiree.EventContext, topic soiree.TypedTopic[T], payload T, extra Fields) *Scope {
	baseCtx := context.Background()
	triggerEvent := ""
	if ctx != nil {
		baseCtx = ctx.Context()
		if ctx.Event() != nil {
			triggerEvent = ctx.Event().Topic()
		}
	}

	opName := OperationName(topic.Name())
	origin := OriginListeners
	fields := lo.Assign(Fields{FieldPayload: payload}, extra)

	if spec, ok := topic.Observability(); ok {
		if spec.Operation != "" {
			opName = OperationName(spec.Operation)
		}
		if spec.Origin != "" {
			origin = Origin(spec.Origin)
		}
		if spec.TriggerFunc != nil {
			triggerEvent = spec.TriggerFunc(ctx, payload)
		}
	}

	return observer.begin(baseCtx, Operation{
		Name:         opName,
		Origin:       origin,
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
