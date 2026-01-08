package soiree

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
)

func TestWorkflowActionCompletedPayloadRoundTrip(t *testing.T) {
	payload := WorkflowActionCompletedPayload{
		InstanceID:  "inst-1",
		ActionIndex: 2,
		ActionType:  enums.WorkflowActionType("0"),
		ObjectID:    "obj-1",
		ObjectType:  "",
		Success:     true,
		Skipped:     true,
	}

	event := NewBaseEvent(TopicWorkflowActionCompleted, payload)

	decoded, err := UnwrapPayload[WorkflowActionCompletedPayload](event)
	if err != nil {
		t.Fatalf("expected payload to unwrap, got error: %v", err)
	}

	if decoded.InstanceID != payload.InstanceID || decoded.Skipped != payload.Skipped {
		t.Fatalf("unexpected payload round trip: %+v", decoded)
	}
}
