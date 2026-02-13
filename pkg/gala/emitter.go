package gala

import "context"

// Emitter exposes the runtime emit API for call sites that should not depend on Runtime directly
type Emitter interface {
	// Emit emits an event payload for one topic
	Emit(context.Context, TopicName, any) EmitReceipt
	// EmitWithHeaders emits an event payload for one topic with explicit headers
	EmitWithHeaders(context.Context, TopicName, any, Headers) EmitReceipt
}
