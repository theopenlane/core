package eventqueue

import "testing"

func TestWorkflowMutationTopicName(t *testing.T) {
	if topic := WorkflowMutationTopicName(" Control "); topic != "workflow.mutation.Control" {
		t.Fatalf("unexpected workflow topic: %q", topic)
	}
}

func TestNotificationMutationTopicName(t *testing.T) {
	if topic := NotificationMutationTopicName(" Task "); topic != "notification.mutation.Task" {
		t.Fatalf("unexpected notification topic: %q", topic)
	}

	if topic := NotificationMutationTopicName(" "); topic != "" {
		t.Fatalf("expected empty topic, got: %q", topic)
	}
}
