package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/gala"
)

func TestEmitWorkflowEventNoRuntime(t *testing.T) {
	receipt := EmitWorkflowEvent(context.Background(), nil, gala.TopicName("test.topic"), map[string]any{"ok": true})
	if receipt.Err != ErrEmitNoEmitter {
		t.Fatalf("expected ErrEmitNoEmitter, got %v", receipt.Err)
	}
	if receipt.Enqueued {
		t.Fatalf("expected Enqueued false when runtime is missing")
	}
}

func TestEmitWorkflowEnvelopeNoRuntime(t *testing.T) {
	receipt := EmitWorkflowEnvelope(context.Background(), nil, gala.Envelope{Topic: gala.TopicName("test.topic")})
	if receipt.Err != ErrEmitNoEmitter {
		t.Fatalf("expected ErrEmitNoEmitter, got %v", receipt.Err)
	}
	if receipt.Enqueued {
		t.Fatalf("expected Enqueued false when runtime is missing")
	}
}

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

func TestParseEmitFailureDetailsEmpty(t *testing.T) {
	_, err := ParseEmitFailureDetails(nil)
	if err != ErrEmitFailureDetailsMissing {
		t.Fatalf("expected ErrEmitFailureDetailsMissing, got %v", err)
	}
}
