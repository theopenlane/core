package runtime

import (
	"context"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestDispatchWebhookEventNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.DispatchWebhookEvent(context.Background(), nil, "hook", types.WebhookReceivedEvent{})
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestDispatchWebhookEventUnknownDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.DispatchWebhookEvent(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "nonexistent",
	}, "hook", types.WebhookReceivedEvent{Name: "push"})
	if err == nil {
		t.Fatal("expected error for unknown definition")
	}
}

func TestEnsureWebhookMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.EnsureWebhook(context.Background(), &ent.Integration{
		DefinitionID: "nonexistent",
	}, "hook", "")
	if !errors.Is(err, registry.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestEnsureWebhookNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.EnsureWebhook(context.Background(), nil, "hook", "")
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestDispatchWebhookEventUnknownWebhookName(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	})

	rt := NewForTesting(reg)
	err := rt.DispatchWebhookEvent(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, "nonexistent-hook", types.WebhookReceivedEvent{Name: "push"})
	if err == nil {
		t.Fatal("expected error for unknown webhook name")
	}
}

func TestEnsureWebhookNameNotRegistered(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	})

	rt := NewForTesting(reg)
	_, err := rt.EnsureWebhook(context.Background(), &ent.Integration{
		DefinitionID: "test-def",
	}, "nonexistent-hook", "")
	if !errors.Is(err, registry.ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound, got %v", err)
	}
}
