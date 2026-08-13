package workflows

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/gala"
)

// EmitReceipt captures enqueue status for workflow events
type EmitReceipt struct {
	// EventID is the event identifier when assigned
	EventID string
	// Err is the enqueue/dispatch error when present
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
	// Topic is the event topic name
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

// EmitWorkflowEvent emits a typed workflow payload via Gala
func EmitWorkflowEvent(ctx context.Context, runtime *gala.Gala, topic gala.TopicName, payload any) EmitReceipt {
	return EmitWorkflowEventWithHeaders(ctx, runtime, topic, payload, gala.Headers{})
}

// EmitWorkflowEventWithHeaders emits a typed workflow payload via Gala with explicit headers
func EmitWorkflowEventWithHeaders(ctx context.Context, runtime *gala.Gala, topic gala.TopicName, payload any, headers gala.Headers) EmitReceipt {
	if runtime == nil {
		return EmitReceipt{
			EventID: string(gala.NewEventID()),
			Err:     ErrEmitNoEmitter,
		}
	}

	id, err := runtime.EmitWithHeaders(ctx, topic, payload, headers)

	return EmitReceipt{
		EventID: string(id),
		Err:     err,
	}
}
