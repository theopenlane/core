package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
)

// credentialTestOAuth is the credential type behind the OAuth test connection
type credentialTestOAuth struct{}

// credentialTestAPIKey is the credential type behind the API key test connection
type credentialTestAPIKey struct{}

// verifyTestClient is the client the verification test connections build
type verifyTestClient struct{}

// verifyTestMetadata is the installation metadata type that derives its own identity
type verifyTestMetadata struct {
	// Tenant is the tenant the credential is scoped to
	Tenant string `json:"tenant"`
}

// InstallationIdentity returns the tenant as the installation identity
func (m verifyTestMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{ExternalID: m.Tenant}
}

// verifyTestPlainMetadata is the installation metadata type that keeps the installation's own identity
type verifyTestPlainMetadata struct{}

var (
	credentialTestOAuthRef  = types.NewConnection[credentialTestOAuth]("credentialTestOAuth")
	credentialTestAPIKeyRef = types.NewConnection[credentialTestAPIKey]("credentialTestAPIKey")
)

// verifyTestRuntime builds a runtime whose keystore can build and pool clients
func verifyTestRuntime(t *testing.T) *Runtime {
	t.Helper()

	store, err := keystore.NewStore(&ent.Client{})
	if err != nil {
		t.Fatalf("expected keystore, got %v", err)
	}

	rt := NewForTesting(registry.New())
	do.ProvideValue(rt.injector, store)

	return rt
}

// verifyTestConnection builds a connection whose verification reports tenant
func verifyTestConnection(tenant string) types.ConnectionRef[credentialTestOAuth] {
	return types.NewConnection[credentialTestOAuth]("credentialTestOAuth").
		Provides(func(context.Context, types.ConnectionRequest[credentialTestOAuth]) (*verifyTestClient, error) {
			return &verifyTestClient{}, nil
		}).
		Verified(func(context.Context, types.ConnectionRequest[credentialTestOAuth], *verifyTestClient) (verifyTestMetadata, error) {
			return verifyTestMetadata{Tenant: tenant}, nil
		})
}

func TestResolveConnectionFromStateEmptyProviderState(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	}

	rt := NewForTesting(registry.New())
	_, found, err := rt.resolveConnectionFromState(def, &ent.Integration{
		DefinitionID: "test-def",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found {
		t.Fatal("expected not found for empty provider state")
	}
}

func TestResolveConnectionFromStateNilProviders(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	}

	rt := NewForTesting(registry.New())
	_, found, err := rt.resolveConnectionFromState(def, &ent.Integration{
		DefinitionID:  "test-def",
		ProviderState: types.IntegrationProviderState{},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found {
		t.Fatal("expected not found for nil providers map")
	}
}

func TestResolveConnectionFromStateWithPersistedRef(t *testing.T) {
	t.Parallel()

	credRef := credentialTestOAuthRef.Connection().Credential.Name
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		Connections:    []types.Connector{credentialTestOAuthRef.Name("OAuth Connection")},
	}

	providerState, _ := json.Marshal(types.DefinitionProviderState{CredentialRef: credRef})

	rt := NewForTesting(registry.New())
	conn, found, err := rt.resolveConnectionFromState(def, &ent.Integration{
		DefinitionID: "test-def",
		ProviderState: types.IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"test-def": providerState,
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !found {
		t.Fatal("expected connection to be found")
	}

	if conn.Name != "OAuth Connection" {
		t.Fatalf("expected OAuth Connection, got %q", conn.Name)
	}
}

func TestResolveConnectionFromStateUnknownRef(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	}

	providerState, _ := json.Marshal(types.DefinitionProviderState{CredentialRef: "unknown"})

	rt := NewForTesting(registry.New())
	_, _, err := rt.resolveConnectionFromState(def, &ent.Integration{
		DefinitionID: "test-def",
		ProviderState: types.IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"test-def": providerState,
			},
		},
	})
	if !errors.Is(err, ErrConnectionNotFound) {
		t.Fatalf("expected ErrConnectionNotFound, got %v", err)
	}
}

func TestDisconnectNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.Disconnect(context.Background(), nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestDisconnectMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.Disconnect(context.Background(), &ent.Integration{
		DefinitionID: "nonexistent",
	})
	if !errors.Is(err, registry.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestReconcileCredentialNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.ReconcileCredential(context.Background(), nil, "", types.CredentialSet{})
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestReconcileCredentialMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.ReconcileCredential(context.Background(), &ent.Integration{
		DefinitionID: "nonexistent",
	}, "", types.CredentialSet{})
	if !errors.Is(err, registry.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestResolvePersistedConnectionSingleConnectionFallback(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	conn, err := rt.resolvePersistedConnection(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		Connections:    []types.Connector{credentialTestOAuthRef.Name("Only Connection")},
	}, &ent.Integration{
		DefinitionID: "test-def",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if conn.Name != "Only Connection" {
		t.Fatalf("expected Only Connection, got %q", conn.Name)
	}
}

func TestResolvePersistedConnectionMultipleConnectionsNoState(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.resolvePersistedConnection(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		Connections: []types.Connector{
			credentialTestOAuthRef.Name("OAuth"),
			credentialTestAPIKeyRef.Name("API Key"),
		},
	}, &ent.Integration{
		DefinitionID: "test-def",
	})
	if !errors.Is(err, ErrConnectionRequired) {
		t.Fatalf("expected ErrConnectionRequired, got %v", err)
	}
}

func TestCheckInstallationInstanceMatchRejectsInstanceMismatch(t *testing.T) {
	t.Parallel()

	installation := &ent.Integration{
		InstallationMetadata: types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: "tenant-a"}},
	}
	metadata := types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: "tenant-b"}}

	err := checkInstallationInstanceMatch(installation, metadata)
	if !errors.Is(err, ErrInstallationInstanceMismatch) {
		t.Fatalf("expected ErrInstallationInstanceMismatch, got %v", err)
	}
}

func TestCheckInstallationInstanceMatchKeepsMatchingInstance(t *testing.T) {
	t.Parallel()

	installation := &ent.Integration{
		InstallationMetadata: types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: "tenant-a"}},
	}
	metadata := types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: "tenant-a"}}

	if err := checkInstallationInstanceMatch(installation, metadata); err != nil {
		t.Fatalf("expected no error for a matching instance, got %v", err)
	}
}

func TestCheckInstallationInstanceMatchAcceptsFirstResolution(t *testing.T) {
	t.Parallel()

	metadata := types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: "tenant-b"}}

	if err := checkInstallationInstanceMatch(&ent.Integration{}, metadata); err != nil {
		t.Fatalf("expected no error for a first resolution, got %v", err)
	}
}

func TestRunConnectionHealthCheckReturnsVerificationIdentity(t *testing.T) {
	t.Parallel()

	rt := verifyTestRuntime(t)
	connection := verifyTestConnection("tenant-b").Connection()
	def := types.Definition{Installation: types.InstallationOf[verifyTestMetadata]().Registration()}

	metadata, err := rt.runConnectionHealthCheck(context.Background(), &ent.Integration{ID: "installation-id"}, def, connection, types.CredentialSet{Data: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if metadata.Display.ExternalID != "tenant-b" {
		t.Fatalf("expected tenant-b, got %q", metadata.Display.ExternalID)
	}
}

func TestRunConnectionHealthCheckRejectsIdentifiableWithoutInstanceID(t *testing.T) {
	t.Parallel()

	rt := verifyTestRuntime(t)
	connection := verifyTestConnection("").Connection()
	def := types.Definition{Installation: types.InstallationOf[verifyTestMetadata]().Registration()}

	_, err := rt.runConnectionHealthCheck(context.Background(), &ent.Integration{ID: "installation-id"}, def, connection, types.CredentialSet{Data: json.RawMessage(`{}`)})
	if !errors.Is(err, ErrInstallationInstanceIDRequired) {
		t.Fatalf("expected ErrInstallationInstanceIDRequired, got %v", err)
	}
}

func TestRunConnectionHealthCheckFallsBackToInstallationIDWhenNotIdentifiable(t *testing.T) {
	t.Parallel()

	rt := verifyTestRuntime(t)
	connection := types.NewConnection[credentialTestOAuth]("credentialTestOAuth").
		Provides(func(context.Context, types.ConnectionRequest[credentialTestOAuth]) (*verifyTestClient, error) {
			return &verifyTestClient{}, nil
		}).
		Verified(func(context.Context, types.ConnectionRequest[credentialTestOAuth], *verifyTestClient) (verifyTestPlainMetadata, error) {
			return verifyTestPlainMetadata{}, nil
		}).Connection()
	def := types.Definition{Installation: types.InstallationOf[verifyTestPlainMetadata]().Registration()}

	metadata, err := rt.runConnectionHealthCheck(context.Background(), &ent.Integration{ID: "installation-id"}, def, connection, types.CredentialSet{Data: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if metadata.Display.ExternalID != "installation-id" {
		t.Fatalf("expected the installation's own id, got %q", metadata.Display.ExternalID)
	}
}
