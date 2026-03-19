package types

import (
	"encoding/json"
	"testing"
)

func TestDefinitionRefID(t *testing.T) {
	t.Parallel()

	ref := NewDefinitionRef("def_01TEST0000000000000000001")
	if ref.ID() != "def_01TEST0000000000000000001" {
		t.Fatalf("DefinitionRef.ID() = %q", ref.ID())
	}
}

func TestClientRefIDsAreDistinct(t *testing.T) {
	t.Parallel()

	first := NewClientRef[string]()
	second := NewClientRef[string]()

	if !first.ID().Valid() {
		t.Fatal("first client ref is not valid")
	}

	if !second.ID().Valid() {
		t.Fatal("second client ref is not valid")
	}

	if first.ID() == second.ID() {
		t.Fatal("client refs should be distinct")
	}

	if first.ID().String() == second.ID().String() {
		t.Fatal("client ref strings should be distinct")
	}
}

func TestOperationRefTopic(t *testing.T) {
	t.Parallel()

	ref := NewOperationRef[struct{}]("health.default")
	if ref.Name() != "health.default" {
		t.Fatalf("OperationRef.Name() = %q", ref.Name())
	}

	if ref.Topic("github_app") != "integration.github_app.health.default" {
		t.Fatalf("OperationRef.Topic() = %q", ref.Topic("github_app"))
	}
}

func TestWebhookRefName(t *testing.T) {
	t.Parallel()

	ref := NewWebhookRef("installation.events")
	if ref.Name() != "installation.events" {
		t.Fatalf("WebhookRef.Name() = %q", ref.Name())
	}
}

func TestClientRefCastSuccess(t *testing.T) {
	t.Parallel()

	ref := NewClientRef[string]()
	got, err := ref.Cast("hello")
	if err != nil {
		t.Fatalf("Cast() error = %v", err)
	}

	if got != "hello" {
		t.Fatalf("Cast() = %q, want %q", got, "hello")
	}
}

func TestClientRefCastFailure(t *testing.T) {
	t.Parallel()

	ref := NewClientRef[string]()
	_, err := ref.Cast(42)
	if err == nil {
		t.Fatal("Cast() expected error, got nil")
	}
}

func TestOperationRefUnmarshalConfig(t *testing.T) {
	t.Parallel()

	type cfg struct {
		Name string `json:"name"`
	}

	ref := NewOperationRef[cfg]("test.op")
	got, err := ref.UnmarshalConfig(json.RawMessage(`{"name":"foo"}`))
	if err != nil {
		t.Fatalf("UnmarshalConfig() error = %v", err)
	}

	if got.Name != "foo" {
		t.Fatalf("UnmarshalConfig() name = %q, want %q", got.Name, "foo")
	}
}

func TestOperationRefUnmarshalConfigNil(t *testing.T) {
	t.Parallel()

	ref := NewOperationRef[struct{}]("test.noop")
	_, err := ref.UnmarshalConfig(nil)
	if err != nil {
		t.Fatalf("UnmarshalConfig(nil) error = %v", err)
	}
}

func TestWebhookEventRefUnmarshalPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Action string `json:"action"`
	}

	ref := NewWebhookEventRef[payload]("created")
	got, err := ref.UnmarshalPayload(json.RawMessage(`{"action":"opened"}`))
	if err != nil {
		t.Fatalf("UnmarshalPayload() error = %v", err)
	}

	if got.Action != "opened" {
		t.Fatalf("UnmarshalPayload() action = %q, want %q", got.Action, "opened")
	}
}
