package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/events/soiree"
)

type testEmitter struct {
	err error
}

// Emit sends a payload and returns an error channel
func (e testEmitter) Emit(topic string, payload any) <-chan error {
	ch := make(chan error, 1)
	if e.err != nil {
		ch <- e.err
	}
	close(ch)
	return ch
}

// TestEmitWorkflowEventWithEventNoEmitter verifies missing emitter handling
func TestEmitWorkflowEventWithEventNoEmitter(t *testing.T) {
	event := soiree.NewBaseEvent("test.topic", json.RawMessage(`{"ok":true}`))
	receipt := EmitWorkflowEventWithEvent(context.Background(), nil, "test.topic", event, nil)
	if receipt.Err != ErrEmitNoEmitter {
		t.Fatalf("expected ErrEmitNoEmitter, got %v", receipt.Err)
	}
	if receipt.Enqueued {
		t.Fatalf("expected Enqueued false when emitter missing")
	}
}

// TestEmitWorkflowEventWithEventAssignsID verifies event id assignment
func TestEmitWorkflowEventWithEventAssignsID(t *testing.T) {
	emitter := testEmitter{}
	event := soiree.NewBaseEvent("test.topic", map[string]any{"ok": true})
	receipt := EmitWorkflowEventWithEvent(context.Background(), emitter, "test.topic", event, nil)
	if receipt.Err != nil {
		t.Fatalf("unexpected error: %v", receipt.Err)
	}
	if receipt.EventID == "" {
		t.Fatalf("expected EventID to be assigned")
	}
	if soiree.EventID(event) == "" {
		t.Fatalf("expected event properties to include EventID")
	}
	if soiree.EventID(event) != receipt.EventID {
		t.Fatalf("expected receipt EventID %s, got %s", receipt.EventID, soiree.EventID(event))
	}
}

// TestEmitWorkflowEventWithEventEnqueueError verifies enqueue error handling
func TestEmitWorkflowEventWithEventEnqueueError(t *testing.T) {
	errEnqueue := errors.New("enqueue failed")
	emitter := testEmitter{err: errEnqueue}
	event := soiree.NewBaseEvent("test.topic", map[string]any{"ok": true})
	receipt := EmitWorkflowEventWithEvent(context.Background(), emitter, "test.topic", event, nil)
	if receipt.Err != errEnqueue {
		t.Fatalf("expected enqueue error, got %v", receipt.Err)
	}
	if receipt.Enqueued {
		t.Fatalf("expected Enqueued false on error")
	}
}

// TestEmitFailureDetailsRoundTrip verifies failure details serialization
func TestEmitFailureDetailsRoundTrip(t *testing.T) {
	meta := EmitFailureMeta{
		EventType:   enums.WorkflowEventTypeActionCompleted,
		ActionKey:   "action",
		ActionIndex: 2,
		ObjectID:    "obj",
		ObjectType:  enums.WorkflowObjectTypeControl,
	}

	details, err := NewEmitFailureDetails("topic", "evt", map[string]any{"ok": true}, meta, errors.New("boom"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.Attempts != 1 {
		t.Fatalf("expected attempts 1, got %d", details.Attempts)
	}
	if details.LastError == "" {
		t.Fatalf("expected LastError to be set")
	}

	encoded, err := json.Marshal(details)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	parsed, err := ParseEmitFailureDetails(encoded)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed.Topic != details.Topic {
		t.Fatalf("expected topic %s, got %s", details.Topic, parsed.Topic)
	}
}

// TestParseEmitFailureDetailsEmpty verifies empty failure details error
func TestParseEmitFailureDetailsEmpty(t *testing.T) {
	_, err := ParseEmitFailureDetails(nil)
	if err != ErrEmitFailureDetailsMissing {
		t.Fatalf("expected ErrEmitFailureDetailsMissing, got %v", err)
	}
}

// TestEmitWorkflowEventNoEmitter verifies typed-topic emission without an emitter
func TestEmitWorkflowEventNoEmitter(t *testing.T) {
	topic := soiree.NewTypedTopic[map[string]any]("test.topic")
	receipt := EmitWorkflowEvent(context.Background(), nil, topic, map[string]any{"ok": true}, nil)
	if receipt.Err != ErrEmitNoEmitter {
		t.Fatalf("expected ErrEmitNoEmitter, got %v", receipt.Err)
	}
	if receipt.Enqueued {
		t.Fatalf("expected Enqueued false when emitter missing")
	}
}

// TestEmitWorkflowEventWrapError verifies wrap failures are returned
func TestEmitWorkflowEventWrapError(t *testing.T) {
	topic := soiree.NewTypedTopic[int]("test.topic", soiree.WithWrap[int](nil))
	receipt := EmitWorkflowEvent(context.Background(), testEmitter{}, topic, 1, nil)
	if receipt.Err == nil {
		t.Fatalf("expected wrap error")
	}
	if receipt.Enqueued {
		t.Fatalf("expected Enqueued false on wrap error")
	}
}
