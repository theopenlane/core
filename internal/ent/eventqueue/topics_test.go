package eventqueue

import "testing"

// TestMutationTopicName verifies concern-specific topic name construction
func TestMutationTopicName(t *testing.T) {
	direct := MutationTopicName(MutationConcernDirect, " Risk ")
	if direct != "Risk" {
		t.Fatalf("unexpected direct topic: %q", direct)
	}

	workflow := MutationTopicName(MutationConcernWorkflow, " Control ")
	if workflow != "workflow.mutation.Control" {
		t.Fatalf("unexpected workflow topic: %q", workflow)
	}

	notification := MutationTopicName(MutationConcernNotification, " Task ")
	if notification != "notification.mutation.Task" {
		t.Fatalf("unexpected notification topic: %q", notification)
	}

	empty := MutationTopicName(MutationConcernNotification, " ")
	if empty != "" {
		t.Fatalf("expected empty topic, got: %q", empty)
	}
}

// TestMutationTopic verifies typed mutation topic construction across concerns
func TestMutationTopic(t *testing.T) {
	topic := MutationTopic(MutationConcernDirect, " Risk ")
	if topic.Name != "Risk" {
		t.Fatalf("unexpected mutation topic: %q", topic.Name)
	}

	empty := MutationTopic(MutationConcernDirect, " ")
	if empty.Name != "" {
		t.Fatalf("expected empty mutation topic name, got: %q", empty.Name)
	}

	workflow := MutationTopic(MutationConcernWorkflow, " Control ")
	if workflow.Name != "workflow.mutation.Control" {
		t.Fatalf("unexpected workflow mutation topic: %q", workflow.Name)
	}

	notification := MutationTopic(MutationConcernNotification, " Task ")
	if notification.Name != "notification.mutation.Task" {
		t.Fatalf("unexpected notification mutation topic: %q", notification.Name)
	}

	emptyWorkflow := MutationTopic(MutationConcernWorkflow, " ")
	if emptyWorkflow.Name != "" {
		t.Fatalf("expected empty workflow mutation topic name, got: %q", emptyWorkflow.Name)
	}
}

// TestMutationTopicUnknownConcernFallsBackToDirect verifies unknown concerns map to direct topics
func TestMutationTopicUnknownConcernFallsBackToDirect(t *testing.T) {
	topic := MutationTopic(MutationConcern("unknown"), " Control ")
	if topic.Name != "Control" {
		t.Fatalf("unexpected topic for unknown concern: %q", topic.Name)
	}
}
