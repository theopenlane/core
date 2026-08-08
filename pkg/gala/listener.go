package gala

import (
	"context"
	"errors"

	"github.com/samber/do/v2"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/jsonx"
)

// ListenerID identifies a registered listener
type ListenerID string

// EventID is a stable identifier used for idempotency and traceability
type EventID string

// NewEventID creates a new event identifier
func NewEventID() EventID {
	return EventID(ulids.New().String())
}

// HandlerContext provides event context and dependency resolution scope for listeners
type HandlerContext struct {
	// Context is the restored event context used for listener execution
	Context context.Context
	// Envelope is the envelope being processed
	Envelope Envelope
	// Injector provides typed dependency lookup via samber/do
	Injector do.Injector
	// Caller is the pre-resolved caller for this dispatch, never nil
	Caller *auth.Caller
}

// Handler processes a typed event payload
type Handler[T any] func(HandlerContext, T) error

// TopicName is the stable string identifier for a topic
type TopicName string

// Topic defines a strongly typed topic contract
type Topic[T any] struct {
	// Name is the stable topic identifier
	Name TopicName
	// UniqueKey optionally derives Headers.UniqueKey from the payload for every emission on
	// the topic; opt out per emission via Headers.SkipUniqueKey
	UniqueKey func(T) string
}

// Definition defines one listener binding
type Definition[T any] struct {
	// Topic is the topic handled by this listener
	Topic Topic[T]
	// Name is the stable listener name; empty defaults to the topic name
	Name string
	// Operations optionally scopes listener interest to specific mutation operations
	// Empty means the listener accepts all operations for the topic
	Operations []string
	// Gate optionally drops the event silently when it returns false
	Gate func(T) bool
	// Caller optionally replaces or augments the restored caller; it receives the restored
	// caller (never nil) and its result is set on the context and HandlerContext.Caller
	Caller func(restored *auth.Caller, payload T) *auth.Caller
	// Elevate optionally transforms the restored context before the handler runs; existing
	// context constructors (privacy allow, capability elevation) plug in unchanged
	Elevate func(context.Context, T) context.Context
	// Enrich optionally derives observability context before the handler runs
	Enrich func(context.Context, T) context.Context
	// Cancel optionally classifies a handler error as terminal, converting it to river.JobCancel
	Cancel func(context.Context, T, error) bool
	// Schedule makes this listener a self-sustaining adaptive re-emit loop when non-nil;
	// exactly one of Handle and Schedule.Handle must be set
	Schedule *ScheduleSpec[T]
	// Handle is the callback invoked for this listener
	Handle Handler[T]
}

// TopicFor derives a durable topic identity from the payload type, naming the topic
// namespace + "." + the payload's JSON schema identifier
func TopicFor[T any](namespace string) Topic[T] {
	return Topic[T]{Name: TopicName(namespace + "." + jsonx.SchemaID(jsonx.SchemaFrom[T]()))}
}

// Register registers listeners on the gala runtime and ensures their topic contracts are configured
func Register[T any](g *Gala, definitions ...Definition[T]) ([]ListenerID, error) {
	if g == nil {
		return nil, ErrGalaRequired
	}

	ids := make([]ListenerID, 0, len(definitions))

	for _, definition := range definitions {
		err := registerTopic(g.registry, definition.Topic, JSONCodec[T]{})
		if err != nil && !errors.Is(err, ErrTopicAlreadyRegistered) {
			return nil, err
		}

		id, err := attachListener(g, definition)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
