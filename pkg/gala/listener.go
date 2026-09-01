package gala

import (
	"context"
	"errors"
	"reflect"

	"github.com/samber/do/v2"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/logx"
)

// ListenerID identifies a registered listener
type ListenerID string

// EventID is a stable identifier for traceability; callers may override the generated value via WithEventID
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
	// Caller is the pre-resolved caller for this dispatch
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
	// Kind is the job kind emissions on this topic dispatch under
	Kind string
	// UniqueKey optionally derives Headers.UniqueKey from the payload for every emission on the topic
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
	// Gate optionally drops the event silently when it returns false; it receives the
	// restored context before any caller or context-key mutation
	Gate func(context.Context, T) bool
	// Caller optionally replaces or augments the restored caller (never nil)
	Caller func(restored *auth.Caller, payload T) *auth.Caller
	// ContextKeys are applied to the context in order before the handler runs
	ContextKeys []func(context.Context) context.Context
	// LogFields is merged over the automatic dispatch log fields
	LogFields func(T) map[string]any
	// Cancel optionally classifies a handler error as terminal, converting it to river.JobCancel
	Cancel func(context.Context, T, error) bool
	// OnExhausted runs when a scheduled loop stops on its error-streak budget
	OnExhausted func(context.Context, T, error)
	// Schedule makes this listener a self-sustaining adaptive re-emit loop when non-nil;
	// exactly one of Handle and Schedule.Handle must be set
	Schedule *ScheduleSpec[T]
	// Handle is the callback invoked for this listener
	Handle Handler[T]
}

// Resolve resolves a listener dependency from the injector, reporting false when the
// dependency is not wired so the listener can skip the event
func Resolve[T any](ctx context.Context, injector do.Injector, listener string) (T, bool) {
	value, err := do.Invoke[T](injector)
	if err != nil || isNilValue(value) {
		logx.FromContext(ctx).Debug().Str("listener", listener).Msg("listener skipped: dependency unresolved")

		var zero T

		return zero, false
	}

	return value, true
}

// isNilValue reports whether a resolved dependency is a typed nil
func isNilValue(value any) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
}

// Registration is one listener registration value: Definition implements it directly and
// declarative config structs implement it by compiling to a Definition
type Registration interface {
	// Attach registers the listener and its topic contract on the runtime
	Attach(g *Gala) (ListenerID, error)
}

// Attach registers the definition's topic contract and listener on the runtime
func (d Definition[T]) Attach(g *Gala) (ListenerID, error) {
	if g == nil {
		return "", ErrGalaRequired
	}

	if err := registerTopic(g.registry, d.Topic); err != nil && !errors.Is(err, ErrTopicAlreadyRegistered) {
		return "", err
	}

	return attachListener(g, d)
}

// Register registers listener values of any payload type through the single registration path
func Register(g *Gala, registrations ...Registration) ([]ListenerID, error) {
	if g == nil {
		return nil, ErrGalaRequired
	}

	ids := make([]ListenerID, 0, len(registrations))

	for _, registration := range registrations {
		id, err := registration.Attach(g)
		if err != nil {
			return ids, err
		}

		// scheduled definitions skip on in-memory pools and return no id
		if id != "" {
			ids = append(ids, id)
		}
	}

	return ids, nil
}
