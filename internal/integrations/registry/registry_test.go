package registry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// testCredentialSlot is a reusable credential slot for tests
var testCredentialSlot = integrationtypes.NewCredentialSlotID("api_key")

// testCredentialSchema is a minimal valid JSON schema for credential registration
var testCredentialSchema = json.RawMessage(`{"type":"object"}`)

// newTestHandler returns a no-op operation handler
func newTestHandler() integrationtypes.OperationHandler {
	return func(context.Context, integrationtypes.OperationRequest) (json.RawMessage, error) {
		return json.RawMessage(`{"ok":true}`), nil
	}
}

// newTestIngestHandler returns a no-op ingest handler
func newTestIngestHandler() integrationtypes.IngestHandler {
	return func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
		return nil, nil
	}
}

// minimalDefinition returns a valid definition with one credential, client, and operation
func minimalDefinition(id string) (integrationtypes.Definition, integrationtypes.ClientRef[string]) {
	clientRef := integrationtypes.NewClientRef[string]()

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          id,
			DisplayName: "Test",
			Active:      true,
			Visible:     true,
		},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: testCredentialSlot, Schema: testCredentialSchema},
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref:            clientRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialSlot},
				Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) {
					return "ok", nil
				},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:      "health.default",
				Topic:     gala.TopicName("integration." + id + ".health.default"),
				ClientRef: clientRef.ID(),
				Handle:    newTestHandler(),
			},
		},
	}, clientRef
}

// TestRegistryRegisterAndResolveDefinition verifies one definition can be registered and resolved
func TestRegistryRegisterAndResolveDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def, clientRef := minimalDefinition("def_resolve")

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	byID, ok := reg.Definition(def.ID)
	if !ok {
		t.Fatalf("Definition() did not find %q", def.ID)
	}

	if byID.DisplayName != def.DisplayName {
		t.Fatalf("Definition() display name = %q, want %q", byID.DisplayName, def.DisplayName)
	}

	client, err := reg.Client(def.ID, clientRef.ID())
	if err != nil {
		t.Fatalf("Client() error = %v", err)
	}

	if client.Ref != clientRef.ID() {
		t.Fatalf("Client() ref mismatch")
	}

	operation, err := reg.Operation(def.ID, "health.default")
	if err != nil {
		t.Fatalf("Operation() error = %v", err)
	}

	if operation.Topic != gala.TopicName("integration.def_resolve.health.default") {
		t.Fatalf("Operation() topic = %q", operation.Topic)
	}

	if got := len(reg.Catalog()); got != 1 {
		t.Fatalf("Catalog() len = %d, want 1", got)
	}

	if got := len(reg.Listeners()); got != 1 {
		t.Fatalf("Listeners() len = %d, want 1", got)
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
				Handle:    newTestHandler(),
			},
			{
				Name:      "second.inspect",
				Topic:     gala.TopicName("integration.multi_client.second.inspect"),
				ClientRef: secondClient.ID(),
				Handle:    newTestHandler(),
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

// TestValidateDefinitionIDRequired verifies empty ID is rejected
func TestValidateDefinitionIDRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{}

	err := reg.Register(def)
	if !errors.Is(err, ErrDefinitionIDRequired) {
		t.Fatalf("expected ErrDefinitionIDRequired, got %v", err)
	}
}

// TestValidateDefinitionAlreadyRegistered verifies duplicate IDs are rejected
func TestValidateDefinitionAlreadyRegistered(t *testing.T) {
	t.Parallel()

	reg := New()
	def, _ := minimalDefinition("def_dup")

	if err := reg.Register(def); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrDefinitionAlreadyRegistered) {
		t.Fatalf("expected ErrDefinitionAlreadyRegistered, got %v", err)
	}
}

// TestValidateOperatorConfigSchemaRequired verifies operator config without schema is rejected
func TestValidateOperatorConfigSchemaRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_opconfig"},
		OperatorConfig: &integrationtypes.OperatorConfigRegistration{Schema: nil},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrOperatorConfigSchemaRequired) {
		t.Fatalf("expected ErrOperatorConfigSchemaRequired, got %v", err)
	}
}

// TestValidateCredentialSchemaRequired verifies credential without schema (non-auth-managed) is rejected
func TestValidateCredentialSchemaRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("orphan")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_credschema"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: nil},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrCredentialSchemaRequired) {
		t.Fatalf("expected ErrCredentialSchemaRequired, got %v", err)
	}
}

// TestValidateCredentialSchemaSkippedForAuthManaged verifies auth-managed credential slots bypass schema requirement
func TestValidateCredentialSchemaSkippedForAuthManaged(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("oauth_token")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_authmanaged"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				Auth: &integrationtypes.AuthRegistration{
					CredentialRef: slot,
				},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   "health",
				Topic:  gala.TopicName("integration.def_authmanaged.health"),
				Handle: newTestHandler(),
			},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() should succeed for auth-managed credential, got %v", err)
	}
}

// TestValidateUserInputSchemaRequired verifies user input without schema is rejected
func TestValidateUserInputSchemaRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_userinput"},
		UserInput:      &integrationtypes.UserInputRegistration{Schema: nil},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrUserInputSchemaRequired) {
		t.Fatalf("expected ErrUserInputSchemaRequired, got %v", err)
	}
}

// TestIndexClientsInvalidRef verifies client with zero-value ref is rejected
func TestIndexClientsInvalidRef(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_badclient"},
		Clients: []integrationtypes.ClientRegistration{
			{Build: func(context.Context, integrationtypes.ClientBuildRequest) (any, error) { return nil, nil }},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrClientRequired) {
		t.Fatalf("expected ErrClientRequired, got %v", err)
	}
}

// TestIndexClientsCredentialRefNotDeclared verifies client referencing undeclared credential is rejected
func TestIndexClientsCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	clientRef := integrationtypes.NewClientRef[string]()
	undeclared := integrationtypes.NewCredentialSlotID("ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_badcredref"},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref:            clientRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{undeclared},
				Build:          func(context.Context, integrationtypes.ClientBuildRequest) (any, error) { return nil, nil },
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrCredentialRefNotDeclared) {
		t.Fatalf("expected ErrCredentialRefNotDeclared, got %v", err)
	}
}

// TestIndexOperationsHandlerRequired verifies operation without any handler is rejected
func TestIndexOperationsHandlerRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_nohandler"},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "bad", Topic: gala.TopicName("bad")},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrOperationHandlerRequired) {
		t.Fatalf("expected ErrOperationHandlerRequired, got %v", err)
	}
}

// TestIndexOperationsHandlerAmbiguous verifies operation with both handlers is rejected
func TestIndexOperationsHandlerAmbiguous(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_ambiguous"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:         "both",
				Topic:        gala.TopicName("both"),
				Handle:       newTestHandler(),
				IngestHandle: newTestIngestHandler(),
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrOperationHandlerAmbiguous) {
		t.Fatalf("expected ErrOperationHandlerAmbiguous, got %v", err)
	}
}

// TestIndexOperationsIngestContractsRequired verifies ingest handler without contracts is rejected
func TestIndexOperationsIngestContractsRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_noingest"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:         "ingest_no_contracts",
				Topic:        gala.TopicName("ingest_no_contracts"),
				IngestHandle: newTestIngestHandler(),
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrIngestContractsRequired) {
		t.Fatalf("expected ErrIngestContractsRequired, got %v", err)
	}
}

// TestIndexOperationsIngestHandlerWithContracts verifies ingest handler with contracts succeeds
func TestIndexOperationsIngestHandlerWithContracts(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_ingest_ok"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:         "ingest_ok",
				Topic:        gala.TopicName("ingest_ok"),
				Ingest:       []integrationtypes.IngestContract{{Schema: "finding"}},
				IngestHandle: newTestIngestHandler(),
			},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

// TestIndexOperationsClientRefNotFound verifies operation referencing unknown client is rejected
func TestIndexOperationsClientRefNotFound(t *testing.T) {
	t.Parallel()

	reg := New()
	ghost := integrationtypes.NewClientRef[string]()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_ghostclient"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:      "bad",
				Topic:     gala.TopicName("bad"),
				ClientRef: ghost.ID(),
				Handle:    newTestHandler(),
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}
}

// TestIndexWebhooksEventResolverRequired verifies webhook with events but no resolver is rejected
func TestIndexWebhooksEventResolverRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_noresolver"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "hooks",
				Events: []integrationtypes.WebhookEventRegistration{
					{
						Name:   "push",
						Topic:  gala.TopicName("push"),
						Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
					},
				},
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrWebhookEventResolverRequired) {
		t.Fatalf("expected ErrWebhookEventResolverRequired, got %v", err)
	}
}

// TestIndexWebhooksEventHandlerRequired verifies webhook event without handler is rejected
func TestIndexWebhooksEventHandlerRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_noevthandler"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "hooks",
				Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
					return integrationtypes.WebhookReceivedEvent{}, nil
				},
				Events: []integrationtypes.WebhookEventRegistration{
					{Name: "push", Topic: gala.TopicName("push")},
				},
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrWebhookEventHandlerRequired) {
		t.Fatalf("expected ErrWebhookEventHandlerRequired, got %v", err)
	}
}

// TestWebhookRegistrationAndLookup verifies webhook registration and all lookup paths
func TestWebhookRegistrationAndLookup(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_webhooks"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "github",
				Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
					return integrationtypes.WebhookReceivedEvent{}, nil
				},
				Events: []integrationtypes.WebhookEventRegistration{
					{
						Name:   "push",
						Topic:  gala.TopicName("integration.def_webhooks.webhook.push"),
						Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
					},
					{
						Name:   "pull_request",
						Topic:  gala.TopicName("integration.def_webhooks.webhook.pr"),
						Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil },
					},
				},
			},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Webhook lookup
	wh, err := reg.Webhook(def.ID, "github")
	if err != nil {
		t.Fatalf("Webhook() error = %v", err)
	}

	if wh.Name != "github" {
		t.Fatalf("Webhook() name = %q, want github", wh.Name)
	}

	// WebhookEvent lookup — happy path
	evt, err := reg.WebhookEvent(def.ID, "github", "push")
	if err != nil {
		t.Fatalf("WebhookEvent() error = %v", err)
	}

	if evt.Name != "push" {
		t.Fatalf("WebhookEvent() name = %q, want push", evt.Name)
	}

	// WebhookEvent lookup — unknown event
	_, err = reg.WebhookEvent(def.ID, "github", "nonexistent")
	if !errors.Is(err, ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound for unknown event, got %v", err)
	}

	// WebhookEvent lookup — unknown webhook name
	_, err = reg.WebhookEvent(def.ID, "nonexistent", "push")
	if !errors.Is(err, ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound for unknown webhook, got %v", err)
	}

	// WebhookEvent lookup — unknown definition
	_, err = reg.WebhookEvent("nonexistent", "github", "push")
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}

	// WebhookListeners returns all events indexed by topic
	listeners := reg.WebhookListeners()
	if got := len(listeners); got != 2 {
		t.Fatalf("WebhookListeners() len = %d, want 2", got)
	}
}

// TestDefinitionNotFound verifies lookups for unknown definition IDs
func TestDefinitionNotFound(t *testing.T) {
	t.Parallel()

	reg := New()

	_, ok := reg.Definition("nonexistent")
	if ok {
		t.Fatal("Definition() should return false for unknown ID")
	}

	_, err := reg.Client("nonexistent", integrationtypes.ClientID{})
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("Client() expected ErrDefinitionNotFound, got %v", err)
	}

	_, err = reg.Operation("nonexistent", "op")
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("Operation() expected ErrDefinitionNotFound, got %v", err)
	}

	_, err = reg.Webhook("nonexistent", "wh")
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("Webhook() expected ErrDefinitionNotFound, got %v", err)
	}
}

// TestClientNotFoundInDefinition verifies client lookup for unknown client ID within a valid definition
func TestClientNotFoundInDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def, _ := minimalDefinition("def_noclient")

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	unknown := integrationtypes.NewClientRef[string]()

	_, err := reg.Client(def.ID, unknown.ID())
	if !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}
}

// TestOperationNotFoundInDefinition verifies operation lookup for unknown name within a valid definition
func TestOperationNotFoundInDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def, _ := minimalDefinition("def_noop")

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err := reg.Operation(def.ID, "nonexistent")
	if !errors.Is(err, ErrOperationNotFound) {
		t.Fatalf("expected ErrOperationNotFound, got %v", err)
	}
}

// TestWebhookNotFoundInDefinition verifies webhook lookup for unknown name within a valid definition
func TestWebhookNotFoundInDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def, _ := minimalDefinition("def_nowh")

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err := reg.Webhook(def.ID, "nonexistent")
	if !errors.Is(err, ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound, got %v", err)
	}
}

// TestDefinitionsReturnsSortedByID verifies Definitions returns entries sorted by definition ID
func TestDefinitionsReturnsSortedByID(t *testing.T) {
	t.Parallel()

	reg := New()
	ids := []string{"def_charlie", "def_alpha", "def_bravo"}

	for _, id := range ids {
		def := integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{ID: id, DisplayName: id},
			Operations: []integrationtypes.OperationRegistration{
				{Name: "h", Topic: gala.TopicName(id + ".h"), Handle: newTestHandler()},
			},
		}
		if err := reg.Register(def); err != nil {
			t.Fatalf("Register(%s) error = %v", id, err)
		}
	}

	defs := reg.Definitions()
	if len(defs) != 3 {
		t.Fatalf("Definitions() len = %d, want 3", len(defs))
	}

	if defs[0].ID != "def_alpha" || defs[1].ID != "def_bravo" || defs[2].ID != "def_charlie" {
		t.Fatalf("Definitions() not sorted: %q, %q, %q", defs[0].ID, defs[1].ID, defs[2].ID)
	}
}

// TestCatalogReturnsSortedByID verifies Catalog returns specs sorted by definition ID
func TestCatalogReturnsSortedByID(t *testing.T) {
	t.Parallel()

	reg := New()
	ids := []string{"def_zulu", "def_mike"}

	for _, id := range ids {
		def := integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{ID: id, DisplayName: id},
			Operations: []integrationtypes.OperationRegistration{
				{Name: "h", Topic: gala.TopicName(id + ".h"), Handle: newTestHandler()},
			},
		}
		if err := reg.Register(def); err != nil {
			t.Fatalf("Register(%s) error = %v", id, err)
		}
	}

	specs := reg.Catalog()
	if len(specs) != 2 {
		t.Fatalf("Catalog() len = %d, want 2", len(specs))
	}

	if specs[0].ID != "def_mike" || specs[1].ID != "def_zulu" {
		t.Fatalf("Catalog() not sorted: %q, %q", specs[0].ID, specs[1].ID)
	}
}

// TestListenersReturnsSortedByTopic verifies Listeners returns operations sorted by topic
func TestListenersReturnsSortedByTopic(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_listeners"},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "b_op", Topic: gala.TopicName("topic.bravo"), Handle: newTestHandler()},
			{Name: "a_op", Topic: gala.TopicName("topic.alpha"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	listeners := reg.Listeners()
	if len(listeners) != 2 {
		t.Fatalf("Listeners() len = %d, want 2", len(listeners))
	}

	if listeners[0].Topic != "topic.alpha" || listeners[1].Topic != "topic.bravo" {
		t.Fatalf("Listeners() not sorted: %q, %q", listeners[0].Topic, listeners[1].Topic)
	}
}

// TestRegisterAllSuccess verifies RegisterAll with valid builders
func TestRegisterAllSuccess(t *testing.T) {
	t.Parallel()

	reg := New()

	b1 := func() (integrationtypes.Definition, error) {
		def, _ := minimalDefinition("def_all_1")
		return def, nil
	}

	b2 := func() (integrationtypes.Definition, error) {
		return integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_all_2", DisplayName: "Two"},
			Operations: []integrationtypes.OperationRegistration{
				{Name: "h", Topic: gala.TopicName("def_all_2.h"), Handle: newTestHandler()},
			},
		}, nil
	}

	if err := reg.RegisterAll(b1, b2); err != nil {
		t.Fatalf("RegisterAll() error = %v", err)
	}

	if got := len(reg.Definitions()); got != 2 {
		t.Fatalf("Definitions() len = %d, want 2", got)
	}
}

// TestRegisterAllNilBuilder verifies RegisterAll rejects nil builder
func TestRegisterAllNilBuilder(t *testing.T) {
	t.Parallel()

	reg := New()

	err := reg.RegisterAll(nil)
	if !errors.Is(err, ErrBuilderNil) {
		t.Fatalf("expected ErrBuilderNil, got %v", err)
	}
}

// TestRegisterAllBuilderError verifies RegisterAll propagates builder errors
func TestRegisterAllBuilderError(t *testing.T) {
	t.Parallel()

	reg := New()
	buildErr := errors.New("build failed")

	b := func() (integrationtypes.Definition, error) {
		return integrationtypes.Definition{}, buildErr
	}

	err := reg.RegisterAll(b)
	if !errors.Is(err, buildErr) {
		t.Fatalf("expected build error, got %v", err)
	}
}

// TestRegisterAllRegistrationError verifies RegisterAll propagates registration errors
func TestRegisterAllRegistrationError(t *testing.T) {
	t.Parallel()

	reg := New()

	b := func() (integrationtypes.Definition, error) {
		return integrationtypes.Definition{DefinitionSpec: integrationtypes.DefinitionSpec{ID: ""}}, nil
	}

	err := reg.RegisterAll(b)
	if !errors.Is(err, ErrDefinitionIDRequired) {
		t.Fatalf("expected ErrDefinitionIDRequired, got %v", err)
	}
}

// TestConnectionCredentialRefRequired verifies connection without credential ref is rejected
func TestConnectionCredentialRefRequired(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_nocred"},
		Connections: []integrationtypes.ConnectionRegistration{
			{},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionCredentialRefRequired) {
		t.Fatalf("expected ErrConnectionCredentialRefRequired, got %v", err)
	}
}

// TestConnectionCredentialRefNotDeclared verifies connection referencing undeclared credential is rejected
func TestConnectionCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	undeclared := integrationtypes.NewCredentialSlotID("ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_undecl"},
		Connections: []integrationtypes.ConnectionRegistration{
			{CredentialRef: undeclared},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionCredentialRefNotDeclared) {
		t.Fatalf("expected ErrConnectionCredentialRefNotDeclared, got %v", err)
	}
}

// TestConnectionAdditionalCredentialRefNotDeclared verifies connection with extra undeclared credential ref is rejected
func TestConnectionAdditionalCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("valid")
	extra := integrationtypes.NewCredentialSlotID("extra_ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_extraref"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot, extra},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionCredentialRefNotDeclared) {
		t.Fatalf("expected ErrConnectionCredentialRefNotDeclared, got %v", err)
	}
}

// TestConnectionClientRefNotDeclared verifies connection referencing undeclared client is rejected
func TestConnectionClientRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("tok")
	ghost := integrationtypes.NewClientRef[string]()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_badclient"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				ClientRefs:     []integrationtypes.ClientID{ghost.ID()},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionClientRefNotDeclared) {
		t.Fatalf("expected ErrConnectionClientRefNotDeclared, got %v", err)
	}
}

// TestConnectionValidationOperationNotDeclared verifies connection with unknown validation operation is rejected
func TestConnectionValidationOperationNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("tok")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_badvalop"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:       slot,
				CredentialRefs:      []integrationtypes.CredentialSlotID{slot},
				ValidationOperation: "nonexistent",
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionValidationOperationNotDeclared) {
		t.Fatalf("expected ErrConnectionValidationOperationNotDeclared, got %v", err)
	}
}

// TestConnectionAuthCredentialRefNotDeclared verifies connection auth with undeclared credential ref is rejected
func TestConnectionAuthCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("tok")
	authSlot := integrationtypes.NewCredentialSlotID("auth_ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_badauth"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				Auth:           &integrationtypes.AuthRegistration{CredentialRef: authSlot},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionAuthCredentialRefNotDeclared) {
		t.Fatalf("expected ErrConnectionAuthCredentialRefNotDeclared, got %v", err)
	}
}

// TestConnectionAuthCredentialRefEmpty verifies connection auth with zero-value credential ref is rejected
func TestConnectionAuthCredentialRefEmpty(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("tok")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_emptyauth"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				Auth:           &integrationtypes.AuthRegistration{},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionAuthCredentialRefNotDeclared) {
		t.Fatalf("expected ErrConnectionAuthCredentialRefNotDeclared, got %v", err)
	}
}

// TestConnectionDisconnectCredentialRefNotDeclared verifies connection disconnect with undeclared credential ref is rejected
func TestConnectionDisconnectCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("tok")
	discSlot := integrationtypes.NewCredentialSlotID("disc_ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_baddisc"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				Disconnect:     &integrationtypes.DisconnectRegistration{CredentialRef: discSlot},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionDisconnectCredentialRefNotDeclared) {
		t.Fatalf("expected ErrConnectionDisconnectCredentialRefNotDeclared, got %v", err)
	}
}

// TestConnectionDisconnectCredentialRefEmpty verifies connection disconnect with zero-value credential ref is rejected
func TestConnectionDisconnectCredentialRefEmpty(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("tok")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_emptydisc"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  slot,
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				Disconnect:     &integrationtypes.DisconnectRegistration{},
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionDisconnectCredentialRefNotDeclared) {
		t.Fatalf("expected ErrConnectionDisconnectCredentialRefNotDeclared, got %v", err)
	}
}

// TestConnectionFullyWiredSuccess verifies a fully wired connection with auth, disconnect, validation, and client refs succeeds
func TestConnectionFullyWiredSuccess(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("primary")
	authSlot := integrationtypes.NewCredentialSlotID("oauth")
	clientRef := integrationtypes.NewClientRef[string]()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_full"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
			{Ref: authSlot},
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref:            clientRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{slot},
				Build:          func(context.Context, integrationtypes.ClientBuildRequest) (any, error) { return "ok", nil },
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "health", Topic: gala.TopicName("health"), Handle: newTestHandler()},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:       slot,
				CredentialRefs:      []integrationtypes.CredentialSlotID{slot, authSlot},
				ClientRefs:          []integrationtypes.ClientID{clientRef.ID()},
				ValidationOperation: "health",
				Auth:                &integrationtypes.AuthRegistration{CredentialRef: authSlot},
				Disconnect:          &integrationtypes.DisconnectRegistration{CredentialRef: slot},
			},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

// TestConnectionAutoAppendsCredentialRef verifies CredentialRef is auto-appended to CredentialRefs when not present
func TestConnectionAutoAppendsCredentialRef(t *testing.T) {
	t.Parallel()

	reg := New()
	slot := integrationtypes.NewCredentialSlotID("auto")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_autoappend"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: slot, Schema: testCredentialSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef: slot,
				// CredentialRefs intentionally empty — should auto-append slot
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

// TestEmptyRegistryLookups verifies all collection methods return empty on fresh registry
func TestEmptyRegistryLookups(t *testing.T) {
	t.Parallel()

	reg := New()

	if got := len(reg.Definitions()); got != 0 {
		t.Fatalf("Definitions() len = %d, want 0", got)
	}

	if got := len(reg.Catalog()); got != 0 {
		t.Fatalf("Catalog() len = %d, want 0", got)
	}

	if got := len(reg.Listeners()); got != 0 {
		t.Fatalf("Listeners() len = %d, want 0", got)
	}

	if got := len(reg.WebhookListeners()); got != 0 {
		t.Fatalf("WebhookListeners() len = %d, want 0", got)
	}
}
