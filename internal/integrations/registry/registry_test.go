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

// testClient is the client type every test connection provides
type testClient struct{}

// testSecondClient is a client type a test connection may leave unprovided
type testSecondClient struct{}

// testMetadata is the installation metadata type every test connection verifies to
type testMetadata struct {
	Tenant string `json:"tenant"`
}

// testOtherMetadata is an installation metadata type differing from the definition's
type testOtherMetadata struct {
	Zone string `json:"zone"`
}

// testConnector yields a fixed connection value
type testConnector integrationtypes.Connection

// Connection returns the fixed connection
func (c testConnector) Connection() integrationtypes.Connection {
	return integrationtypes.Connection(c)
}

var (
	// testRuntimeSchema is the reflected runtime config schema behind the runtime integration tests
	testRuntimeSchema = jsonx.SchemaFrom[testRuntimeConfig]()
	// testClientRef is the name of the client every test connection provides
	testClientRef = testConnectionOf[testCredential]().Connection().Verify.ClientRef
)

// testConnectionOf returns a connection over credential T that provides and verifies the test client
func testConnectionOf[T any]() integrationtypes.ConnectionRef[T] {
	return integrationtypes.ConnectionOf[T]().
		Provides(func(context.Context, integrationtypes.ConnectionRequest[T]) (*testClient, error) {
			return &testClient{}, nil
		}).
		Verified(func(context.Context, integrationtypes.ConnectionRequest[T], *testClient) (testMetadata, error) {
			return testMetadata{}, nil
		})
}

// newTestInstallation returns the installation metadata layout every test connection verifies to
func newTestInstallation() *integrationtypes.InstallationRegistration {
	return integrationtypes.InstallationOf[testMetadata]().Registration()
}

// connectionDefinition returns a definition declaring the given connections beside the test installation layout
func connectionDefinition(id string, connections ...integrationtypes.Connector) integrationtypes.Definition {
	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: id},
		Connections:    connections,
		Installation:   newTestInstallation(),
	}
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

// minimalDefinition returns a valid definition with one connection, installation layout, and operation
func minimalDefinition(id string) integrationtypes.Definition {
	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          id,
			DisplayName: "Test",
			Active:      true,
			Visible:     true,
		},
		Connections:  []integrationtypes.Connector{testConnectionOf[testCredential]()},
		Installation: newTestInstallation(),
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:      "health.default",
				Topic:     gala.TopicName("integration." + id + ".health.default"),
				ClientRef: testClientRef,
				Handle:    newTestHandler(),
			},
		},
	}
}

// TestRegistryRegisterAndResolveDefinition verifies one definition can be registered and resolved
func TestRegistryRegisterAndResolveDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def := minimalDefinition("def_resolve")

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

	operation, err := reg.Operation(def.ID, "health.default")
	if err != nil {
		t.Fatalf("Operation() error = %v", err)
	}

	if operation.Topic != gala.TopicName("integration.run.def_resolve.health.default") {
		t.Fatalf("Operation() topic = %q", operation.Topic)
	}

	if got := len(reg.Catalog()); got != 1 {
		t.Fatalf("Catalog() len = %d, want 1", got)
	}

	if got := len(reg.Listeners()); got != 1 {
		t.Fatalf("Listeners() len = %d, want 1", got)
	}
}

// secondClientOperation returns an operation running under the second test client
func secondClientOperation(name string) integrationtypes.OperationRegistration {
	return integrationtypes.NewOperationPayload[finalizePayload](name).
		Handles(func(context.Context, integrationtypes.OperationRequest, *testSecondClient, finalizePayload) (json.RawMessage, error) {
			return nil, nil
		}).
		Registration()
}

// TestRegistrySupportsMultipleClientsPerConnection verifies a connection providing several clients serves operations under each
func TestRegistrySupportsMultipleClientsPerConnection(t *testing.T) {
	t.Parallel()

	reg := New()

	connection := testConnectionOf[testCredential]().
		Provides(func(context.Context, integrationtypes.ConnectionRequest[testCredential]) (*testSecondClient, error) {
			return &testSecondClient{}, nil
		})

	definition := minimalDefinition("def_multi_client")
	definition.Connections = []integrationtypes.Connector{connection}
	definition.Operations = append(definition.Operations, secondClientOperation("second.inspect"))

	if err := reg.Register(definition); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for _, name := range []string{"health.default", "second.inspect"} {
		if _, err := reg.Operation(definition.ID, name); err != nil {
			t.Fatalf("Operation(%s) error = %v", name, err)
		}
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
	def := minimalDefinition("def_dup")

	if err := reg.Register(def); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrDefinitionAlreadyRegistered) {
		t.Fatalf("expected ErrDefinitionAlreadyRegistered, got %v", err)
	}
}

// assertDuplicateRejected verifies registration fails with ErrDuplicateRegistration
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

// TestDuplicateConnectionSlotRejected verifies same-slot connections are rejected
func TestDuplicateConnectionSlotRejected(t *testing.T) {
	t.Parallel()

	reg := New()
	def := connectionDefinition("def_dup_conn", testConnectionOf[testCredential](), testConnectionOf[testCredential]())

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

// TestDuplicateWebhookEventNameRejected verifies same-named events in one webhook are rejected
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

// TestValidateCredentialSchemaRequired verifies a credential without a stored schema is rejected
func TestValidateCredentialSchemaRequired(t *testing.T) {
	t.Parallel()

	reg := New()

	orphan := testConnectionOf[testCredential]().Connection()
	orphan.Credential.Schema = nil

	def := connectionDefinition("def_credschema", testConnector(orphan))

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
		UserInput:      &integrationtypes.InputRegistration{Name: "UserInput"},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrUserInputSchemaRequired) {
		t.Fatalf("expected ErrUserInputSchemaRequired, got %v", err)
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

// TestIndexOperationsSnapshotRequiresIngestHandle verifies Snapshot needs an IngestHandle
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

// TestIndexOperationsSnapshotWithIngestHandle verifies Snapshot with an IngestHandle succeeds
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

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_ghostclient"},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:      "bad",
				Topic:     gala.TopicName("bad"),
				ClientRef: testClientRef,
				Handle:    newTestHandler(),
			},
		},
	}

	err := reg.Register(def)
	if !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}
}

// TestIndexOperationsConnectionClientNotProvided verifies an operation's client missing from a connection is rejected
func TestIndexOperationsConnectionClientNotProvided(t *testing.T) {
	t.Parallel()

	reg := New()

	def := minimalDefinition("def_unprovided_client")
	def.Operations = append(def.Operations, secondClientOperation("second.inspect"))

	err := reg.Register(def)
	if !errors.Is(err, ErrConnectionClientNotProvided) {
		t.Fatalf("expected ErrConnectionClientNotProvided, got %v", err)
	}
}

// TestIndexOperationsClientConflict verifies an operation whose handlers bound different clients is rejected
func TestIndexOperationsClientConflict(t *testing.T) {
	t.Parallel()

	reg := New()

	def := minimalDefinition("def_client_conflict")
	def.Operations[0].ClientConflict = "other"

	err := reg.Register(def)
	if !errors.Is(err, ErrOperationClientConflict) {
		t.Fatalf("expected ErrOperationClientConflict, got %v", err)
	}
}

// TestIndexOperationsClientRefWithRuntimeIntegration verifies a runtime integration serves an operation naming a client without connections
func TestIndexOperationsClientRefWithRuntimeIntegration(t *testing.T) {
	t.Parallel()

	reg := New()

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_runtime_clientref"},
		RuntimeIntegration: &integrationtypes.RuntimeIntegrationRegistration{
			Schema: testRuntimeSchema,
			Build:  func(_ context.Context, _ json.RawMessage) (any, error) { return nil, nil },
		},
		Operations: []integrationtypes.OperationRegistration{
			{Name: "op", Topic: gala.TopicName("op"), ClientRef: testClientRef, Handle: newTestHandler()},
		},
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
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

	_, err := reg.Operation("nonexistent", "op")
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("Operation() expected ErrDefinitionNotFound, got %v", err)
	}

	_, err = reg.Webhook("nonexistent", "wh")
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("Webhook() expected ErrDefinitionNotFound, got %v", err)
	}
}

// TestOperationNotFoundInDefinition verifies lookup of an unknown operation name fails
func TestOperationNotFoundInDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def := minimalDefinition("def_noop")

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err := reg.Operation(def.ID, "nonexistent")
	if !errors.Is(err, ErrOperationNotFound) {
		t.Fatalf("expected ErrOperationNotFound, got %v", err)
	}
}

// TestWebhookNotFoundInDefinition verifies lookup of an unknown webhook name fails
func TestWebhookNotFoundInDefinition(t *testing.T) {
	t.Parallel()

	reg := New()
	def := minimalDefinition("def_nowh")

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

	if listeners[0].Topic != "integration.run.def_listeners.a_op" || listeners[1].Topic != "integration.run.def_listeners.b_op" {
		t.Fatalf("Listeners() not sorted: %q, %q", listeners[0].Topic, listeners[1].Topic)
	}
}

// TestRegisterAllSuccess verifies RegisterAll with valid builders
func TestRegisterAllSuccess(t *testing.T) {
	t.Parallel()

	reg := New()

	b1 := func() (integrationtypes.Definition, error) {
		return minimalDefinition("def_all_1"), nil
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

// TestConnectionVerifyRequired verifies a connection without a verification is rejected
func TestConnectionVerifyRequired(t *testing.T) {
	t.Parallel()

	reg := New()

	unverified := integrationtypes.ConnectionOf[testCredential]().
		Provides(func(context.Context, integrationtypes.ConnectionRequest[testCredential]) (*testClient, error) {
			return &testClient{}, nil
		})

	err := reg.Register(connectionDefinition("def_conn_noverify", unverified))
	if !errors.Is(err, ErrConnectionVerifyRequired) {
		t.Fatalf("expected ErrConnectionVerifyRequired, got %v", err)
	}
}

// TestConnectionVerifyClientNotProvided verifies a verification client outside the connection's clients is rejected
func TestConnectionVerifyClientNotProvided(t *testing.T) {
	t.Parallel()

	reg := New()

	misrouted := integrationtypes.ConnectionOf[testCredential]().
		Provides(func(context.Context, integrationtypes.ConnectionRequest[testCredential]) (*testClient, error) {
			return &testClient{}, nil
		}).
		Verified(func(context.Context, integrationtypes.ConnectionRequest[testCredential], *testSecondClient) (testMetadata, error) {
			return testMetadata{}, nil
		})

	err := reg.Register(connectionDefinition("def_conn_verifyclient", misrouted))
	if !errors.Is(err, ErrConnectionVerifyClientNotProvided) {
		t.Fatalf("expected ErrConnectionVerifyClientNotProvided, got %v", err)
	}
}

// TestInstallationRequiredWithConnections verifies connections need an installation metadata layout
func TestInstallationRequiredWithConnections(t *testing.T) {
	t.Parallel()

	reg := New()

	def := connectionDefinition("def_conn_noinstallation", testConnectionOf[testCredential]())
	def.Installation = nil

	err := reg.Register(def)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

// TestInstallationSchemaMismatch verifies a verification returning another metadata layout is rejected
func TestInstallationSchemaMismatch(t *testing.T) {
	t.Parallel()

	reg := New()

	mismatched := integrationtypes.ConnectionOf[testCredential]().
		Provides(func(context.Context, integrationtypes.ConnectionRequest[testCredential]) (*testClient, error) {
			return &testClient{}, nil
		}).
		Verified(func(context.Context, integrationtypes.ConnectionRequest[testCredential], *testClient) (testOtherMetadata, error) {
			return testOtherMetadata{}, nil
		})

	err := reg.Register(connectionDefinition("def_conn_mismatch", mismatched))
	if !errors.Is(err, ErrInstallationSchemaMismatch) {
		t.Fatalf("expected ErrInstallationSchemaMismatch, got %v", err)
	}
}

// TestConnectionFullyWiredSuccess verifies a fully wired connection succeeds
func TestConnectionFullyWiredSuccess(t *testing.T) {
	t.Parallel()

	reg := New()

	def := minimalDefinition("def_conn_full")
	def.Connections = []integrationtypes.Connector{
		testConnectionOf[testCredential]().Disconnects("remove the app", nil),
		testConnectionOf[testAuthCredential]().Authenticates(integrationtypes.NewAuthFlow[testAuthCredential](nil, nil)),
	}

	if err := reg.Register(def); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

// TestRuntimeIntegrationRegistration verifies RuntimeIntegration registers and caches a client
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

// TestRuntimeIntegrationNilConfig verifies nil config registers without a cached client
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
		Connections:  []integrationtypes.Connector{testConnectionOf[testCredential]()},
		Installation: newTestInstallation(),
		UserInput: &integrationtypes.InputRegistration{
			Name:   "UserInput",
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

// TestRuntimeCoexistsWithOperatorConfig verifies runtime integration and operator config coexist
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
	def := minimalDefinition("def_standard")

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

	def := minimalDefinition("static-def")
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

	def := minimalDefinition("no-static")
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
