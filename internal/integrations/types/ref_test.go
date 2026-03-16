package types

import "testing"

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
