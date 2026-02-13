package gala

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"
)

func TestRiverDispatchArgsInsertOptsUsesQueueFallback(t *testing.T) {
	withQueue := RiverDispatchArgs{Queue: "queue_custom"}
	if withQueue.InsertOpts().Queue != "queue_custom" {
		t.Fatalf("expected explicit queue override")
	}

	defaultQueue := RiverDispatchArgs{}
	if defaultQueue.InsertOpts().Queue != DefaultQueueName {
		t.Fatalf("expected default queue %q, got %q", DefaultQueueName, defaultQueue.InsertOpts().Queue)
	}
}

func TestRiverDispatchArgsDecodeEnvelopeErrors(t *testing.T) {
	_, err := (RiverDispatchArgs{}).DecodeEnvelope()
	if !errors.Is(err, ErrRiverDispatchJobEnvelopeRequired) {
		t.Fatalf("expected ErrRiverDispatchJobEnvelopeRequired, got %v", err)
	}

	_, err = (RiverDispatchArgs{Envelope: []byte("{bad")}).DecodeEnvelope()
	if !errors.Is(err, ErrRiverEnvelopeDecodeFailed) {
		t.Fatalf("expected ErrRiverEnvelopeDecodeFailed, got %v", err)
	}
}

func TestRiverDispatcherQueueSelection(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := NewRiverDispatcher(RiverDispatcherOptions{
		JobClient:    client,
		DefaultQueue: "queue_custom_default",
		QueueByClass: map[QueueClass]string{
			QueueClassGeneral: "",
		},
	})
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:            NewEventID(),
		Topic:         TopicName("gala.test.queue_selection"),
		SchemaVersion: 1,
		Payload:       []byte(`{"message":"hello"}`),
	}

	err = dispatcher.DispatchDurable(context.Background(), envelope, TopicPolicy{QueueName: "queue_topic_policy"})
	if err != nil {
		t.Fatalf("unexpected queue-name dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_topic_policy" {
		t.Fatalf("expected topic queue override, got %#v", client.lastOpts)
	}

	err = dispatcher.DispatchDurable(context.Background(), envelope, TopicPolicy{QueueClass: QueueClassGeneral})
	if err != nil {
		t.Fatalf("unexpected fallback dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_custom_default" {
		t.Fatalf("expected custom default queue, got %#v", client.lastOpts)
	}
}

func TestRiverDispatcherQueueSelectionUsesCustomDefaultForBuiltInClass(t *testing.T) {
	client := &riverTestInsertClient{}
	dispatcher, err := NewRiverDispatcher(RiverDispatcherOptions{
		JobClient:    client,
		DefaultQueue: "queue_custom_default",
	})
	if err != nil {
		t.Fatalf("failed to build dispatcher: %v", err)
	}

	envelope := Envelope{
		ID:            NewEventID(),
		Topic:         TopicName("gala.test.queue_builtin_default"),
		SchemaVersion: 1,
		Payload:       []byte(`{"message":"hello"}`),
	}

	err = dispatcher.DispatchDurable(context.Background(), envelope, TopicPolicy{QueueClass: QueueClassWorkflow})
	if err != nil {
		t.Fatalf("unexpected dispatch error: %v", err)
	}

	if client.lastOpts == nil || client.lastOpts.Queue != "queue_custom_default" {
		t.Fatalf("expected custom default queue for built-in class, got %#v", client.lastOpts)
	}
}

func TestRiverDispatchWorkerRequiresRuntimeInstance(t *testing.T) {
	worker := NewRiverDispatchWorker(func() *Runtime {
		return nil
	})

	err := worker.Work(context.Background(), &river.Job[RiverDispatchArgs]{})
	if !errors.Is(err, ErrRuntimeRequired) {
		t.Fatalf("expected ErrRuntimeRequired, got %v", err)
	}
}
