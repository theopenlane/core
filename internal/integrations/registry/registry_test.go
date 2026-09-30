package registry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// testCredential is the credential type behind the reusable test slot
type testCredential struct{}

// testAuthCredential is the credential type behind the auth-managed test slot
type testAuthCredential struct{}

// testRuntimeConfig is the runtime config type behind the runtime integration tests
type testRuntimeConfig struct {
	Key string `json:"key"`
}

var (
	// testCredentialRef is the reusable credential slot for tests
	testCredentialRef = integrationtypes.CredentialRefOf[testCredential]()
	// testAuthCredentialRef is the credential slot an auth flow fills in tests
	testAuthCredentialRef = integrationtypes.CredentialRefOf[testAuthCredential]()
	// testRuntimeSchema is the reflected runtime config schema behind the runtime integration tests
	testRuntimeSchema = jsonx.SchemaFrom[testRuntimeConfig]()
)

// testCredentialRegistration is the reusable credential registration declaring the test slot with its stored schema
var testCredentialRegistration = testCredentialRef.Registration(integrationtypes.CredentialRegistration{})

// newTestHealthCheck returns a definition health check that runs without a client
func newTestHealthCheck() *integrationtypes.HealthCheckRegistration {
	return &integrationtypes.HealthCheckRegistration{Handle: newTestHandler()}
}

// testOperationConfig is the operation config type behind the typed operation ref tests
type testOperationConfig struct {
	Limit int `json:"limit"`
}

// testUserInput is the user input type behind the typed user input ref tests
type testUserInput struct {
	Region string `json:"region"`
}

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
	clientRef := integrationtypes.ClientRefOf[string]()

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          id,
			DisplayName: "Test",
			Active:      true,
			Visible:     true,
		},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref:            clientRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
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
	firstClient := integrationtypes.NewClientRef[string]("first")
	secondClient := integrationtypes.NewClientRef[int]("second")

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

// assertDuplicateRejected verifies registration fails with ErrDuplicateRegistration and leaves the definition unregistered
func assertDuplicateRejected(t *testing.T, reg *Registry, def integrationtypes.Definition) {
	t.Helper()

	err := reg.Register(def)
	if !errors.Is(err, ErrDuplicateRegistration) {
		t.Fatalf("expected ErrDuplicateRegistration, got %v", err)
	}

	if _, ok := reg.Definition(def.ID); ok {
		t.Fatalf("Definition(%q) present after rejected registration", def.ID)
	}
}

// TestDuplicateConnectionSlotRejected verifies two connections on the same credential slot are rejected
func TestDuplicateConnectionSlotRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_dup_conn"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{CredentialRef: testCredentialRef.ID()},
			{CredentialRef: testCredentialRef.ID()},
		},
		HealthCheck: newTestHealthCheck(),
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("def_dup_conn.h"), Handle: newTestHandler()},
		},
	}

	assertDuplicateRejected(t, reg, def)
}

// TestDuplicateClientRejected verifies two clients with the same ref are rejected
func TestDuplicateClientRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	clientRef := integrationtypes.ClientRefOf[string]()
	build := func(context.Context, integrationtypes.ClientBuildRequest) (any, error) { return "ok", nil }

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_dup_client"},
		Clients: []integrationtypes.ClientRegistration{
			{Ref: clientRef.ID(), Build: build},
			{Ref: clientRef.ID(), Build: build},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("def_dup_client.h"), Handle: newTestHandler()},
		},
	}

	assertDuplicateRejected(t, reg, def)
}

// TestDuplicateOperationNameRejected verifies two operations with the same name are rejected
func TestDuplicateOperationNameRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_dup_op"},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "sync", Topic: gala.TopicName("def_dup_op.sync.a"), Handle: newTestHandler()},
			{Name: "sync", Topic: gala.TopicName("def_dup_op.sync.b"), Handle: newTestHandler()},
		},
	}

	assertDuplicateRejected(t, reg, def)
}

// TestDuplicateWebhookNameRejected verifies two webhook contracts with the same name are rejected
func TestDuplicateWebhookNameRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_dup_webhook"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{Name: "hooks"},
			{Name: "hooks"},
		},
	}

	assertDuplicateRejected(t, reg, def)
}

// TestDuplicateWebhookEventNameRejected verifies two events with the same name within one webhook are rejected
func TestDuplicateWebhookEventNameRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	handle := func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil }

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_dup_event"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "hooks",
				Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
					return integrationtypes.WebhookReceivedEvent{}, nil
				},
				Events: []integrationtypes.WebhookEventRegistration{
					{Name: "push", Topic: gala.TopicName("def_dup_event.push.a"), Handle: handle},
					{Name: "push", Topic: gala.TopicName("def_dup_event.push.b"), Handle: handle},
				},
			},
		},
	}

	assertDuplicateRejected(t, reg, def)
}

// TestDuplicateOperationTopicAcrossDefinitionsRejected verifies a second definition claiming a held operation topic is rejected without touching the registry
func TestDuplicateOperationTopicAcrossDefinitionsRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	topic := gala.TopicName("integration.shared.sync")

	first := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_topic_first"},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "sync", Topic: topic, Handle: newTestHandler()},
		},
	}

	if err := reg.Register(first); err != nil {
		t.Fatalf("Register(first) error = %v", err)
	}

	second := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_topic_second"},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "other", Topic: topic, Handle: newTestHandler()},
		},
	}

	assertDuplicateRejected(t, reg, second)

	listeners := reg.Listeners()
	if len(listeners) != 1 || listeners[0].Name != "sync" {
		t.Fatalf("Listeners() = %+v, want only the first definition's operation", listeners)
	}
}

// TestDuplicateWebhookEventTopicAcrossDefinitionsRejected verifies a second definition claiming a held webhook event topic is rejected without touching the registry
func TestDuplicateWebhookEventTopicAcrossDefinitionsRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	topic := gala.TopicName("integration.webhook.shared.push")

	first := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_webhook_topic_first"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "hooks",
				Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
					return integrationtypes.WebhookReceivedEvent{}, nil
				},
				Events: []integrationtypes.WebhookEventRegistration{
					{Name: "push", Topic: topic, Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil }},
				},
			},
		},
	}

	if err := reg.Register(first); err != nil {
		t.Fatalf("Register(first) error = %v", err)
	}

	second := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_webhook_topic_second"},
		Webhooks: []integrationtypes.WebhookRegistration{
			{
				Name: "hooks",
				Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
					return integrationtypes.WebhookReceivedEvent{}, nil
				},
				Events: []integrationtypes.WebhookEventRegistration{
					{Name: "other", Topic: topic, Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error { return nil }},
				},
			},
		},
	}

	assertDuplicateRejected(t, reg, second)

	listeners := reg.WebhookListeners()
	if len(listeners) != 1 || listeners[0].Name != "push" {
		t.Fatalf("WebhookListeners() = %+v, want only the first definition's event", listeners)
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

// TestValidateCredentialSchemaRequired verifies a credential registration without a stored schema is rejected
func TestValidateCredentialSchemaRequired(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_credschema"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			{Ref: testCredentialRef.ID(), Name: "Orphan"},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrCredentialSchemaRequired) {
		t.Fatalf("expected ErrCredentialSchemaRequired, got %v", err)
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
	clientRef := integrationtypes.ClientRefOf[string]()
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

// TestIndexOperationsSnapshotRequiresIngestHandle verifies Policy.Snapshot without an IngestHandle is rejected
func TestIndexOperationsSnapshotRequiresIngestHandle(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_badsnapshot"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   "bad_snapshot",
				Topic:  gala.TopicName("bad_snapshot"),
				Handle: newTestHandler(),
				Policy: integrationtypes.ExecutionPolicy{Snapshot: true},
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrIngestSnapshotRequiresIngestHandle) {
		t.Fatalf("expected ErrIngestSnapshotRequiresIngestHandle, got %v", err)
	}
}

// TestIndexOperationsSnapshotWithIngestHandle verifies Policy.Snapshot with an IngestHandle succeeds
func TestIndexOperationsSnapshotWithIngestHandle(t *testing.T) {
	t.Parallel()

	reg := New()
	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_goodsnapshot"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:         "good_snapshot",
				Topic:        gala.TopicName("good_snapshot"),
				Ingest:       []integrationtypes.IngestContract{{Schema: "directory_account"}},
				IngestHandle: newTestIngestHandler(),
				Policy:       integrationtypes.ExecutionPolicy{Snapshot: true},
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
	ghost := integrationtypes.ClientRefOf[string]()

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

	wh, err := reg.Webhook(def.ID, "github")
	if err != nil {
		t.Fatalf("Webhook() error = %v", err)
	}

	if wh.Name != "github" {
		t.Fatalf("Webhook() name = %q, want github", wh.Name)
	}

	evt, err := reg.WebhookEvent(def.ID, "github", "push")
	if err != nil {
		t.Fatalf("WebhookEvent() error = %v", err)
	}

	if evt.Name != "push" {
		t.Fatalf("WebhookEvent() name = %q, want push", evt.Name)
	}

	_, err = reg.WebhookEvent(def.ID, "github", "nonexistent")
	if !errors.Is(err, ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound for unknown event, got %v", err)
	}

	_, err = reg.WebhookEvent(def.ID, "nonexistent", "push")
	if !errors.Is(err, ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound for unknown webhook, got %v", err)
	}

	_, err = reg.WebhookEvent("nonexistent", "github", "push")
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}

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

	_, err := reg.Client("nonexistent", integrationtypes.NewClientID("nonexistent"))
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

	unknown := integrationtypes.NewClientRef[string]("unknown")

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
		HealthCheck: newTestHealthCheck(),
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
		HealthCheck: newTestHealthCheck(),
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
	extra := integrationtypes.NewCredentialSlotID("extra_ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_extraref"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID(), extra},
			},
		},
		HealthCheck: newTestHealthCheck(),
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
	ghost := integrationtypes.ClientRefOf[string]()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_badclient"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				ClientRefs:     []integrationtypes.ClientID{ghost.ID()},
			},
		},
		HealthCheck: newTestHealthCheck(),
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionClientRefNotDeclared) {
		t.Fatalf("expected ErrConnectionClientRefNotDeclared, got %v", err)
	}
}

// TestHealthCheckRequiredWithConnections verifies a definition declaring connections without a health check is rejected
func TestHealthCheckRequiredWithConnections(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_nohealth"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{CredentialRef: testCredentialRef.ID()},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrHealthCheckRequired) {
		t.Fatalf("expected ErrHealthCheckRequired, got %v", err)
	}
}

// TestHealthCheckHandlerRequired verifies a definition health check without a handler is rejected
func TestHealthCheckHandlerRequired(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_nohealthhandler"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{CredentialRef: testCredentialRef.ID()},
		},
		HealthCheck: &integrationtypes.HealthCheckRegistration{},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionHealthCheckHandlerRequired) {
		t.Fatalf("expected ErrConnectionHealthCheckHandlerRequired, got %v", err)
	}
}

// TestHealthCheckClientNotDeclared verifies a definition health check with an unknown client ref is rejected
func TestHealthCheckClientNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	unknownClient := integrationtypes.ClientRefOf[string]()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_badhealthclient"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{CredentialRef: testCredentialRef.ID()},
		},
		HealthCheck: &integrationtypes.HealthCheckRegistration{
			ClientRef: unknownClient.ID(),
			Handle:    newTestHandler(),
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

// TestHealthCheckClientCredentialMissing verifies a health check client not built from every connection's credential slot is rejected
func TestHealthCheckClientCredentialMissing(t *testing.T) {
	t.Parallel()

	reg := New()
	clientRef := integrationtypes.ClientRefOf[string]().Using(testCredentialRef)

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_health_uncovered"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
			testAuthCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
		},
		Clients: []integrationtypes.ClientRegistration{
			clientRef.Registration(func(context.Context, integrationtypes.ClientBuildRequest) (string, error) { return "ok", nil }, integrationtypes.ClientRegistration{}),
		},
		Connections: []integrationtypes.ConnectionRegistration{
			integrationtypes.NewConnectionRef(testCredentialRef).Enables(clientRef).Registration(integrationtypes.ConnectionRegistration{}),
			integrationtypes.NewConnectionRef(testAuthCredentialRef).Enables(clientRef).Registration(integrationtypes.ConnectionRegistration{}),
		},
		HealthCheck: clientRef.HealthCheck(func(context.Context, integrationtypes.OperationRequest, string) (json.RawMessage, error) {
			return nil, nil
		}),
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("def_health_uncovered.h"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrHealthCheckClientCredentialMissing) {
		t.Fatalf("expected ErrHealthCheckClientCredentialMissing, got %v", err)
	}

	covered := def
	covered.DefinitionSpec.ID = "def_health_covered"
	covered.Connections = covered.Connections[:1]
	covered.Operations = []integrationtypes.OperationRegistration{
		{Name: "h", Topic: gala.TopicName("def_health_covered.h"), Handle: newTestHandler()},
	}

	if err := reg.Register(covered); err != nil {
		t.Fatalf("Register(covered) error = %v", err)
	}
}

// TestConnectionAuthCredentialRefNotDeclared verifies connection auth with undeclared credential ref is rejected
func TestConnectionAuthCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	reg := New()
	authSlot := integrationtypes.NewCredentialSlotID("auth_ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_badauth"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				Auth:           &integrationtypes.AuthRegistration{CredentialRef: authSlot},
			},
		},
		HealthCheck: newTestHealthCheck(),
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

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_emptyauth"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				Auth:           &integrationtypes.AuthRegistration{},
			},
		},
		HealthCheck: newTestHealthCheck(),
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
	discSlot := integrationtypes.NewCredentialSlotID("disc_ghost")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_baddisc"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				Disconnect:     &integrationtypes.DisconnectRegistration{CredentialRef: discSlot},
			},
		},
		HealthCheck: newTestHealthCheck(),
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

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_emptydisc"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				Disconnect:     &integrationtypes.DisconnectRegistration{},
			},
		},
		HealthCheck: newTestHealthCheck(),
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
	clientRef := integrationtypes.ClientRefOf[string]()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_full"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
			testAuthCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref:            clientRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				Build:          func(context.Context, integrationtypes.ClientBuildRequest) (any, error) { return "ok", nil },
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "health", Topic: gala.TopicName("health"), Handle: newTestHandler()},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID(), testAuthCredentialRef.ID()},
				ClientRefs:     []integrationtypes.ClientID{clientRef.ID()},
				Auth:           &integrationtypes.AuthRegistration{CredentialRef: testAuthCredentialRef.ID()},
				Disconnect:     &integrationtypes.DisconnectRegistration{CredentialRef: testCredentialRef.ID()},
			},
		},
		HealthCheck: &integrationtypes.HealthCheckRegistration{
			ClientRef: clientRef.ID(),
			Handle:    newTestHandler(),
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

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_conn_autoappend"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef: testCredentialRef.ID(),
			},
		},
		HealthCheck: newTestHealthCheck(),
		Operations: []integrationtypes.OperationRegistration{
			{Name: "h", Topic: gala.TopicName("h"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

// TestRuntimeIntegrationRegistration verifies a definition with RuntimeIntegration can register and cache a client
func TestRuntimeIntegrationRegistration(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_runtime",
			DisplayName: "Runtime Test",
			Active:      true,
			Visible:     true,
		},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
			Config: json.RawMessage(`{"key":"val"}`),
			Build: func(_ context.Context, config json.RawMessage) (any, error) {
				return "runtime-client-" + string(config), nil
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "send", Topic: gala.TopicName("def_runtime.send"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	client, ok := reg.RuntimeClient("def_runtime")
	if !ok {
		t.Fatal("RuntimeClient() returned false, want true")
	}

	got, _ := client.(string)
	if got != `runtime-client-{"key":"val"}` {
		t.Fatalf("RuntimeClient() = %q, want cached client", got)
	}

	if !reg.IsRuntimeIntegration("def_runtime") {
		t.Fatal("IsRuntimeIntegration() = false, want true")
	}
}

// TestRuntimeIntegrationNilConfig verifies a runtime definition with nil config registers but has no cached client
func TestRuntimeIntegrationNilConfig(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          "def_runtime_nocfg",
			DisplayName: "Runtime No Config",
			Active:      true,
		},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
			Build: func(_ context.Context, _ json.RawMessage) (any, error) {
				return "should-not-be-called", nil
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "op", Topic: gala.TopicName("def_runtime_nocfg.op"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, ok := reg.RuntimeClient("def_runtime_nocfg")
	if ok {
		t.Fatal("RuntimeClient() returned true for nil config, want false")
	}

	if !reg.IsRuntimeIntegration("def_runtime_nocfg") {
		t.Fatal("IsRuntimeIntegration() = false, want true")
	}
}

// TestRuntimeCoexistsWithCredentials verifies runtime integration can coexist with credentials
func TestRuntimeCoexistsWithCredentials(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_runtime_creds", Active: true, Visible: true},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
			Config: json.RawMessage(`{"key":"val"}`),
			Build: func(_ context.Context, config json.RawMessage) (any, error) {
				return "runtime-client", nil
			},
		},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRegistration,
		},
		Clients: []integrationtypes.ClientRegistration{
			{
				Ref:            integrationtypes.ClientRefOf[string]().ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
				Build: func(_ context.Context, _ integrationtypes.ClientBuildRequest) (any, error) {
					return "customer-client", nil
				},
			},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef:  testCredentialRef.ID(),
				CredentialRefs: []integrationtypes.CredentialSlotID{testCredentialRef.ID()},
			},
		},
		HealthCheck: newTestHealthCheck(),
		UserInput: &integrationtypes.UserInputRegistration{
			Schema: json.RawMessage(`{"type":"object"}`),
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "op", Topic: gala.TopicName("def_runtime_creds.op"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	client, ok := reg.RuntimeClient("def_runtime_creds")
	if !ok {
		t.Fatal("RuntimeClient() returned false, want true")
	}

	got, _ := client.(string)
	if got != "runtime-client" {
		t.Fatalf("RuntimeClient() = %q, want %q", got, "runtime-client")
	}

	if !reg.IsRuntimeIntegration("def_runtime_creds") {
		t.Fatal("IsRuntimeIntegration() = false, want true")
	}
}

// TestRuntimeCoexistsWithOperatorConfig verifies a definition can declare both a runtime integration and operator config
func TestRuntimeCoexistsWithOperatorConfig(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_runtime_opconf"},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
			Build:  func(_ context.Context, _ json.RawMessage) (any, error) { return nil, nil },
		},
		OperatorConfig: &integrationtypes.OperatorConfigRegistration{Schema: json.RawMessage(`{"type":"object"}`)},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "op", Topic: gala.TopicName("op"), Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("expected registration to succeed, got %v", err)
	}
}

// TestRuntimeBuildRequired verifies runtime integration rejects nil Build function
func TestRuntimeBuildRequired(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_runtime_nobuild"},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "op", Topic: gala.TopicName("op"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrRuntimeBuildRequired) {
		t.Fatalf("expected ErrRuntimeBuildRequired, got %v", err)
	}
}

// TestRuntimeBuildError verifies build failure during registration propagates
func TestRuntimeBuildError(t *testing.T) {
	t.Parallel()

	reg := New()
	buildErr := errors.New("provider init failed")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_runtime_buildfail"},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
			Config: json.RawMessage(`{}`),
			Build: func(_ context.Context, _ json.RawMessage) (any, error) {
				return nil, buildErr
			},
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "op", Topic: gala.TopicName("op"), Handle: newTestHandler()},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, buildErr) {
		t.Fatalf("expected build error, got %v", err)
	}
}

// TestRuntimeClientMiss verifies RuntimeClient returns false for non-runtime definitions
func TestRuntimeClientMiss(t *testing.T) {
	t.Parallel()

	reg := New()
	def, _ := minimalDefinition("def_standard")

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, ok := reg.RuntimeClient("def_standard")
	if ok {
		t.Fatal("RuntimeClient() returned true for standard definition, want false")
	}

	if reg.IsRuntimeIntegration("def_standard") {
		t.Fatal("IsRuntimeIntegration() = true for standard definition, want false")
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

func TestStaticWebhooks_ReturnsStaticRouteEntries(t *testing.T) {
	t.Parallel()

	reg := New()

	def, _ := minimalDefinition("static-def")
	def.Webhooks = []integrationtypes.WebhookRegistration{
		{
			Name:        "delivery",
			StaticRoute: "/email/webhook",
			Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
				return integrationtypes.WebhookReceivedEvent{}, nil
			},
			Events: []integrationtypes.WebhookEventRegistration{
				{
					Name:  "email.delivered",
					Topic: gala.TopicName("test.email.delivered"),
					Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error {
						return nil
					},
				},
			},
		},
		{
			Name: "dynamic",
			Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
				return integrationtypes.WebhookReceivedEvent{}, nil
			},
			Events: []integrationtypes.WebhookEventRegistration{
				{
					Name:  "push",
					Topic: gala.TopicName("test.push"),
					Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error {
						return nil
					},
				},
			},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	entries := reg.StaticWebhooks()
	if len(entries) != 1 {
		t.Fatalf("StaticWebhooks() len = %d, want 1", len(entries))
	}

	if entries[0].DefinitionID != "static-def" {
		t.Fatalf("DefinitionID = %q, want %q", entries[0].DefinitionID, "static-def")
	}

	if entries[0].WebhookName != "delivery" {
		t.Fatalf("WebhookName = %q, want %q", entries[0].WebhookName, "delivery")
	}

	if entries[0].StaticRoute != "/email/webhook" {
		t.Fatalf("StaticRoute = %q, want %q", entries[0].StaticRoute, "/email/webhook")
	}
}

func TestStaticWebhooks_EmptyWhenNoStaticRoutes(t *testing.T) {
	t.Parallel()

	reg := New()

	def, _ := minimalDefinition("no-static")
	def.Webhooks = []integrationtypes.WebhookRegistration{
		{
			Name: "dynamic-only",
			Event: func(integrationtypes.WebhookInboundRequest) (integrationtypes.WebhookReceivedEvent, error) {
				return integrationtypes.WebhookReceivedEvent{}, nil
			},
			Events: []integrationtypes.WebhookEventRegistration{
				{
					Name:  "push",
					Topic: gala.TopicName("test.push"),
					Handle: func(context.Context, integrationtypes.WebhookHandleRequest) error {
						return nil
					},
				},
			},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	entries := reg.StaticWebhooks()
	if len(entries) != 0 {
		t.Fatalf("StaticWebhooks() len = %d, want 0", len(entries))
	}
}
