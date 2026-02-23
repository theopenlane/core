package observability

import "context"

// warnWithFields logs a workflow warning with the provided fields
func warnWithFields(ctx context.Context, op OperationName, origin Origin, triggerEvent string, fields Fields, err error) {
	warn(ctx, Operation{
		Name:         op,
		Origin:       origin,
		TriggerEvent: triggerEvent,
	}, fields, err)
}

// WarnEngine logs a workflow warning for engine operations (with the trigger event pre-set reducing overhead for callers)
func WarnEngine(ctx context.Context, op OperationName, triggerEvent string, fields Fields, err error) {
	warnWithFields(ctx, op, OriginEngine, triggerEvent, fields, err)
}

// WarnListener logs a workflow warning for listener operations (with the trigger event pre-set reducing overhead for callers)
func WarnListener(ctx context.Context, op OperationName, triggerEvent string, fields Fields, err error) {
	warnWithFields(ctx, op, OriginListeners, triggerEvent, fields, err)
}
