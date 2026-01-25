package workflows

import (
	"context"
	"encoding/json"
	"time"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const emitEnqueueTimeout = 200 * time.Millisecond

// EmitReceipt captures enqueue status for workflow events
type EmitReceipt struct {
	// EventID is the event idempotency key when assigned
	EventID string
	// Enqueued reports whether the event was enqueued
	Enqueued bool
	// Err is the enqueue error when present
	Err error
}

// EmitFailureMeta holds structured emit metadata for recovery
type EmitFailureMeta struct {
	// EventType is the workflow event type associated with the emit
	EventType enums.WorkflowEventType
	// ActionKey is the workflow action key when available
	ActionKey string
	// ActionIndex is the workflow action index when available
	ActionIndex int
	// ObjectID is the target object id
	ObjectID string
	// ObjectType is the target object type
	ObjectType enums.WorkflowObjectType
}

// EmitFailureDetails stores failure context for deterministic re-emission
type EmitFailureDetails struct {
	// Topic is the soiree topic name
	Topic string `json:"topic"`
	// EventID is the idempotency key when available
	EventID string `json:"event_id,omitempty"`
	// Payload is the original event payload JSON
	Payload json.RawMessage `json:"payload,omitempty"`
	// Attempts is the number of emit attempts recorded
	Attempts int `json:"attempts"`
	// LastError is the most recent enqueue error
	LastError string `json:"last_error,omitempty"`
	// OriginalEventType is the workflow event type that failed
	OriginalEventType enums.WorkflowEventType `json:"original_event_type"`
	// ActionKey is the workflow action key when available
	ActionKey string `json:"action_key,omitempty"`
	// ActionIndex is the workflow action index when available
	ActionIndex int `json:"action_index"`
	// ObjectID is the workflow object id
	ObjectID string `json:"object_id,omitempty"`
	// ObjectType is the workflow object type
	ObjectType enums.WorkflowObjectType `json:"object_type,omitempty"`
}

// NewEmitFailureDetails builds a failure payload, including a JSON-encoded event payload
func NewEmitFailureDetails(topic, eventID string, payload any, meta EmitFailureMeta, err error) (EmitFailureDetails, error) {
	encoded, encodeErr := json.Marshal(payload)
	if encodeErr != nil {
		return EmitFailureDetails{}, encodeErr
	}

	details := EmitFailureDetails{
		Topic:             topic,
		EventID:           eventID,
		Payload:           encoded,
		Attempts:          1,
		OriginalEventType: meta.EventType,
		ActionKey:         meta.ActionKey,
		ActionIndex:       meta.ActionIndex,
		ObjectID:          meta.ObjectID,
		ObjectType:        meta.ObjectType,
	}

	if err != nil {
		details.LastError = err.Error()
	}

	return details, nil
}

// ParseEmitFailureDetails decodes failure details from raw JSON
func ParseEmitFailureDetails(raw json.RawMessage) (EmitFailureDetails, error) {
	if len(raw) == 0 {
		return EmitFailureDetails{}, ErrEmitFailureDetailsMissing
	}

	var details EmitFailureDetails
	if err := json.Unmarshal(raw, &details); err != nil {
		return details, err
	}

	return details, nil
}

// EmitWorkflowEvent wraps and emits a typed workflow event, returning enqueue status
func EmitWorkflowEvent[T any](ctx context.Context, emitter soiree.Emitter, topic soiree.TypedTopic[T], payload T, client any) EmitReceipt {
	event, err := topic.Wrap(payload)
	if err != nil {
		return EmitReceipt{Err: err}
	}

	return EmitWorkflowEventWithEvent(ctx, emitter, topic.Name(), event, client)
}

// EmitWorkflowEventWithEvent emits a pre-built event and returns enqueue status
func EmitWorkflowEventWithEvent(ctx context.Context, emitter soiree.Emitter, topic string, event soiree.Event, client any) EmitReceipt {
	if emitter == nil {
		return EmitReceipt{Err: ErrEmitNoEmitter}
	}

	if ctx != nil {
		event.SetContext(ctx)
	}
	if client != nil {
		event.SetClient(client)
	}

	eventID := soiree.EventID(event)
	if eventID == "" {
		props := event.Properties()
		if props == nil {
			props = soiree.NewProperties()
		}
		eventID = ulids.New().String()
		props[soiree.PropertyEventID] = eventID
		event.SetProperties(props)
	}

	errCh := emitter.Emit(topic, event)
	enqueueErr := drainEmitErrors(errCh, emitEnqueueTimeout)

	return EmitReceipt{
		EventID:  eventID,
		Enqueued: enqueueErr == nil,
		Err:      enqueueErr,
	}
}

// drainEmitErrors reads errors until close or timeout and returns the first error
func drainEmitErrors(errCh <-chan error, timeout time.Duration) error {
	if errCh == nil {
		return nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	var firstErr error
	for {
		select {
		case err, ok := <-errCh:
			if !ok {
				return firstErr
			}
			if err != nil && firstErr == nil {
				firstErr = err
			}
		case <-timer.C:
			return firstErr
		}
	}
}
