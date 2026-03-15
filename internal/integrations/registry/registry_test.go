package registry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	generated "github.com/theopenlane/core/internal/ent/generated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// TestRegistryRegisterAndResolveDefinition verifies one definition can be registered and resolved
func TestRegistryRegisterAndResolveDefinition(t *testing.T) {
	t.Parallel()

	reg := New()

	definition := integrationtypes.Definition{
		Spec: integrationtypes.DefinitionSpec{
			ID:          integrationtypes.DefinitionID("def_01HZY6PZQK2T64B7J9QX4N5Z6A"),
			Slug:        "github_app",
			Version:     "v1",
			Family:      "github",
			DisplayName: "GitHub App",
			Active:      true,
			Visible:     true,
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Name: "rest",
				Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) {
					return "ok", nil
				},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  "health.default",
				Topic: gala.TopicName("integration.github_app.health.default"),
				Handle: func(context.Context, *generated.Integration, integrationtypes.CredentialSet, any, json.RawMessage) (json.RawMessage, error) {
					return json.RawMessage(`{"ok":true}`), nil
				},
			},
		},
	}

	if err := reg.Register(definition); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	byID, ok := reg.Definition(definition.Spec.ID)
	if !ok {
		t.Fatalf("Definition() did not find %q", definition.Spec.ID)
	}

	if byID.Spec.DisplayName != definition.Spec.DisplayName {
		t.Fatalf("Definition() display name = %q, want %q", byID.Spec.DisplayName, definition.Spec.DisplayName)
	}

	client, err := reg.Client(definition.Spec.ID, "rest")
	if err != nil {
		t.Fatalf("Client() error = %v", err)
	}

	if client.Name != "rest" {
		t.Fatalf("Client() name = %q, want rest", client.Name)
	}

	operation, err := reg.Operation(definition.Spec.ID, "health.default")
	if err != nil {
		t.Fatalf("Operation() error = %v", err)
	}

	if operation.Topic != gala.TopicName("integration.github_app.health.default") {
		t.Fatalf("Operation() topic = %q", operation.Topic)
	}

	if got := len(reg.Catalog()); got != 1 {
		t.Fatalf("Catalog() len = %d, want 1", got)
	}

	if got := len(reg.Listeners()); got != 1 {
		t.Fatalf("Listeners() len = %d, want 1", got)
	}
}

// TestRegistryRejectsDuplicateOperationTopic verifies operation topics remain unique across definitions
func TestRegistryRejectsDuplicateOperationTopic(t *testing.T) {
	t.Parallel()

	reg := New()
	first := integrationtypes.Definition{
		Spec: integrationtypes.DefinitionSpec{
			ID:          integrationtypes.DefinitionID("def_first"),
			Slug:        "first",
			Version:     "v1",
			DisplayName: "First",
			Active:      true,
			Visible:     true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  "health.default",
				Topic: gala.TopicName("integration.shared.topic"),
			},
		},
	}
	second := integrationtypes.Definition{
		Spec: integrationtypes.DefinitionSpec{
			ID:          integrationtypes.DefinitionID("def_second"),
			Slug:        "second",
			Version:     "v1",
			DisplayName: "Second",
			Active:      true,
			Visible:     true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  "collect.default",
				Topic: gala.TopicName("integration.shared.topic"),
			},
		},
	}

	if err := reg.Register(first); err != nil {
		t.Fatalf("Register(first) error = %v", err)
	}

	err := reg.Register(second)
	if !errors.Is(err, ErrOperationTopicAlreadyRegistered) {
		t.Fatalf("Register(second) error = %v, want %v", err, ErrOperationTopicAlreadyRegistered)
	}
}

func TestRegistryRejectsDuplicateSlug(t *testing.T) {
	t.Parallel()

	reg := New()

	first := integrationtypes.Definition{
		Spec: integrationtypes.DefinitionSpec{
			ID:          "def_first",
			Slug:        "shared_slug",
			Version:     "v1",
			DisplayName: "First",
			Active:      true,
			Visible:     true,
		},
	}
	second := integrationtypes.Definition{
		Spec: integrationtypes.DefinitionSpec{
			ID:          "def_second",
			Slug:        "shared_slug",
			Version:     "v1",
			DisplayName: "Second",
			Active:      true,
			Visible:     true,
		},
	}

	if err := reg.Register(first); err != nil {
		t.Fatalf("Register(first) error = %v", err)
	}

	err := reg.Register(second)
	if !errors.Is(err, ErrDefinitionSlugAlreadyRegistered) {
		t.Fatalf("Register(second) error = %v, want %v", err, ErrDefinitionSlugAlreadyRegistered)
	}
}
