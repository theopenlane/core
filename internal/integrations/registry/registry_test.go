package registry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// TestRegistryRegisterAndResolveDefinition verifies one definition can be registered and resolved
func TestRegistryRegisterAndResolveDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	clientRef := integrationtypes.NewClientRef[string]()

	definition := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_01HZY6PZQK2T64B7J9QX4N5Z6A",
			Slug:        "github_app",
			Family:      "github",
			DisplayName: "GitHub App",
			Active:      true,
			Visible:     true,
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref: clientRef.ID(),
				Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) {
					return "ok", nil
				},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  "health.default",
				Topic: gala.TopicName("integration.github_app.health.default"),
				Handle: func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
					return json.RawMessage(`{"ok":true}`), nil
				},
			},
		},
	}

	if err := reg.Register(definition); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	byID, ok := reg.Definition(definition.ID)
	if !ok {
		t.Fatalf("Definition() did not find %q", definition.ID)
	}

	if byID.DisplayName != definition.DisplayName {
		t.Fatalf("Definition() display name = %q, want %q", byID.DisplayName, definition.DisplayName)
	}

	client, err := reg.Client(definition.ID, clientRef.ID())
	if err != nil {
		t.Fatalf("Client() error = %v", err)
	}

	if client.Ref != clientRef.ID() {
		t.Fatalf("Client() ref mismatch")
	}

	operation, err := reg.Operation(definition.ID, "health.default")
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
	nopHandle := func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
		return nil, nil
	}

	first := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_first",
			Slug:        "first",
			DisplayName: "First",
			Active:      true,
			Visible:     true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   "health.default",
				Topic:  gala.TopicName("integration.shared.topic"),
				Handle: nopHandle,
			},
		},
	}
	second := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_second",
			Slug:        "second",
			DisplayName: "Second",
			Active:      true,
			Visible:     true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   "collect.default",
				Topic:  gala.TopicName("integration.shared.topic"),
				Handle: nopHandle,
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

// TestRegistryRejectsDuplicateOperationTopicSameName verifies topics remain unique even when operation names match
func TestRegistryRejectsDuplicateOperationTopicSameName(t *testing.T) {
	t.Parallel()

	reg := New()
	nopHandle := func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
		return nil, nil
	}

	first := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_first_same_name",
			Slug:        "first_same_name",
			DisplayName: "First",
			Active:      true,
			Visible:     true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   "health.default",
				Topic:  gala.TopicName("integration.shared.same_name_topic"),
				Handle: nopHandle,
			},
		},
	}
	second := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_second_same_name",
			Slug:        "second_same_name",
			DisplayName: "Second",
			Active:      true,
			Visible:     true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   "health.default",
				Topic:  gala.TopicName("integration.shared.same_name_topic"),
				Handle: nopHandle,
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

// TestRegistrySupportsMultipleClientsPerDefinition verifies a definition can register more than one client
func TestRegistrySupportsMultipleClientsPerDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	firstClient := integrationtypes.NewClientRef[string]()
	secondClient := integrationtypes.NewClientRef[int]()

	definition := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_multi_client",
			Slug:        "multi_client",
			DisplayName: "Multi Client",
			Active:      true,
			Visible:     true,
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref: firstClient.ID(),
				Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) {
					return "first", nil
				},
			},
			{
				Ref: secondClient.ID(),
				Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) {
					return 2, nil
				},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:      "first.inspect",
				Topic:     gala.TopicName("integration.multi_client.first.inspect"),
				ClientRef: firstClient.ID(),
				Handle: func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
			},
			{
				Name:      "second.inspect",
				Topic:     gala.TopicName("integration.multi_client.second.inspect"),
				ClientRef: secondClient.ID(),
				Handle: func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
					return nil, nil
				},
			},
		},
	}

	if err := reg.Register(definition); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	firstRegistration, err := reg.Client(definition.ID, firstClient.ID())
	if err != nil {
		t.Fatalf("Client(first) error = %v", err)
	}

	secondRegistration, err := reg.Client(definition.ID, secondClient.ID())
	if err != nil {
		t.Fatalf("Client(second) error = %v", err)
	}

	if firstRegistration.Ref != firstClient.ID() {
		t.Fatalf("Client(first) ref mismatch")
	}

	if secondRegistration.Ref != secondClient.ID() {
		t.Fatalf("Client(second) ref mismatch")
	}
}

// TestRegistryRejectsDuplicateWebhookEventTopicWithinDefinition verifies webhook event topics are unique inside one definition
func TestRegistryRejectsDuplicateWebhookEventTopicWithinDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	definition := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_webhook_topic_collision",
			Slug:        "webhook_topic_collision",
			DisplayName: "Webhook Topic Collision",
			Active:      true,
			Visible:     true,
		},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "default",
				Event: func(context.Context, integrationtypes.WebhookEventRequest) (integrationtypes.WebhookReceivedEvent, error) {
					return integrationtypes.WebhookReceivedEvent{Name: "created"}, nil
				},
				Events: []integrationtypes.WebhookEventRegistration{
					{
						Name:  "created",
						Topic: gala.TopicName("integration.webhook.shared"),
						Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error {
							return nil
						},
					},
					{
						Name:  "updated",
						Topic: gala.TopicName("integration.webhook.shared"),
						Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error {
							return nil
						},
					},
				},
			},
		},
	}

	err := reg.Register(definition)
	if !errors.Is(err, ErrOperationTopicAlreadyRegistered) {
		t.Fatalf("Register() error = %v, want %v", err, ErrOperationTopicAlreadyRegistered)
	}
}

// TestRegistryRejectsDuplicateSlug verifies slugs remain unique across registered definitions
func TestRegistryRejectsDuplicateSlug(t *testing.T) {
	t.Parallel()

	reg := New()

	first := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_first",
			Slug:        "shared_slug",
			DisplayName: "First",
			Active:      true,
			Visible:     true,
		},
	}
	second := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_second",
			Slug:        "shared_slug",
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
