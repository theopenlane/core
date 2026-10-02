package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// credentialTestOAuth is the credential type behind the OAuth test slot
type credentialTestOAuth struct{}

// credentialTestAPIKey is the credential type behind the API key test slot
type credentialTestAPIKey struct{}

var (
	credentialTestOAuthSchema  = jsonx.SchemaFrom[credentialTestOAuth]()
	credentialTestOAuthRef     = types.NewCredentialRef[credentialTestOAuth]("credentialTestOAuth")
	credentialTestAPIKeySchema = jsonx.SchemaFrom[credentialTestAPIKey]()
	credentialTestAPIKeyRef    = types.NewCredentialRef[credentialTestAPIKey]("credentialTestAPIKey")
)

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

	credRef := credentialTestOAuthRef.ID()
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credentialTestOAuthRef.ID(), Name: "OAuth", Schema: credentialTestOAuthSchema},
		},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "OAuth Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
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

	credRef := types.NewCredentialSlotID("unknown")
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	}

	providerState, _ := json.Marshal(types.DefinitionProviderState{CredentialRef: credRef})

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

func TestResolveConnectionForCredentialFromPersistedState(t *testing.T) {
	t.Parallel()

	credRef := credentialTestOAuthRef.ID()
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credentialTestOAuthRef.ID(), Name: "OAuth", Schema: credentialTestOAuthSchema},
		},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "OAuth Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
	}

	providerState, _ := json.Marshal(types.DefinitionProviderState{CredentialRef: credRef})
	installation := &ent.Integration{
		DefinitionID: "test-def",
		ProviderState: types.IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"test-def": providerState,
			},
		},
	}

	rt := NewForTesting(registry.New())
	conn, err := rt.resolveConnectionForCredential(def, installation, credRef)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if conn.Name != "OAuth Connection" {
		t.Fatalf("expected OAuth Connection, got %q", conn.Name)
	}
}

func TestResolveConnectionForCredentialRefNotDeclared(t *testing.T) {
	t.Parallel()

	credRef := credentialTestOAuthRef.ID()
	otherRef := credentialTestAPIKeyRef.ID()
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credentialTestOAuthRef.ID(), Name: "OAuth", Schema: credentialTestOAuthSchema},
		},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "OAuth Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
	}

	providerState, _ := json.Marshal(types.DefinitionProviderState{CredentialRef: credRef})
	installation := &ent.Integration{
		DefinitionID: "test-def",
		ProviderState: types.IntegrationProviderState{
			Providers: map[string]json.RawMessage{
				"test-def": providerState,
			},
		},
	}

	rt := NewForTesting(registry.New())
	_, err := rt.resolveConnectionForCredential(def, installation, otherRef)
	if !errors.Is(err, ErrCredentialNotDeclared) {
		t.Fatalf("expected ErrCredentialNotDeclared, got %v", err)
	}
}

func TestResolveConnectionForCredentialNoStateWithRef(t *testing.T) {
	t.Parallel()

	credRef := credentialTestAPIKeyRef.ID()
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credentialTestAPIKeyRef.ID(), Name: "API Key", Schema: credentialTestAPIKeySchema},
		},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "API Key Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
	}

	installation := &ent.Integration{
		DefinitionID: "test-def",
	}

	rt := NewForTesting(registry.New())
	conn, err := rt.resolveConnectionForCredential(def, installation, credRef)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if conn.Name != "API Key Connection" {
		t.Fatalf("expected API Key Connection, got %q", conn.Name)
	}
}

func TestResolveConnectionForCredentialNoStateEmptyRef(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	}
	installation := &ent.Integration{
		DefinitionID: "test-def",
	}

	rt := NewForTesting(registry.New())
	_, err := rt.resolveConnectionForCredential(def, installation, types.CredentialSlotID{})
	if !errors.Is(err, ErrConnectionRequired) {
		t.Fatalf("expected ErrConnectionRequired, got %v", err)
	}
}

func TestResolveConnectionForCredentialNoStateUnknownRef(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	}
	installation := &ent.Integration{
		DefinitionID: "test-def",
	}

	rt := NewForTesting(registry.New())
	_, err := rt.resolveConnectionForCredential(def, installation, types.NewCredentialSlotID("missing"))
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

func TestReconcileNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.Reconcile(context.Background(), nil, nil, nil, types.CredentialSlotID{}, nil, nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestReconcileMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.Reconcile(context.Background(), &ent.Integration{
		DefinitionID: "nonexistent",
	}, nil, nil, types.CredentialSlotID{}, nil, nil)
	if !errors.Is(err, registry.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestReconcileNoInputNoCredential(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	})

	rt := NewForTesting(reg)
	err := rt.Reconcile(context.Background(), &ent.Integration{
		DefinitionID:      "test-def",
		DefinitionVersion: reg.Version("test-def"),
	}, nil, nil, types.CredentialSlotID{}, nil, nil)
	if err != nil {
		t.Fatalf("expected no error for no-op reconcile, got %v", err)
	}
}

func TestReconcileEmptyInputNoCredential(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	})

	rt := NewForTesting(reg)
	err := rt.Reconcile(context.Background(), &ent.Integration{
		DefinitionID:      "test-def",
		DefinitionVersion: reg.Version("test-def"),
	}, json.RawMessage(`null`), nil, types.CredentialSlotID{}, nil, nil)
	if err != nil {
		t.Fatalf("expected no error for null input reconcile, got %v", err)
	}
}

func TestResolvePersistedConnectionSingleConnectionFallback(t *testing.T) {
	t.Parallel()

	credRef := credentialTestOAuthRef.ID()
	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credentialTestOAuthRef.ID(), Name: "OAuth", Schema: credentialTestOAuthSchema},
		},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "Only Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
		HealthCheck: types.CredentialHealthCheck(func(context.Context, types.OperationRequest) (json.RawMessage, error) { return nil, nil }),
	})

	rt := NewForTesting(reg)
	conn, err := rt.resolvePersistedConnection(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "Only Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
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

	refA := credentialTestOAuthRef.ID()
	refB := credentialTestAPIKeyRef.ID()

	rt := NewForTesting(registry.New())
	_, err := rt.resolvePersistedConnection(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: refA, Name: "OAuth"},
			{CredentialRef: refB, Name: "API Key"},
		},
	}, &ent.Integration{
		DefinitionID: "test-def",
	})
	if !errors.Is(err, ErrConnectionRequired) {
		t.Fatalf("expected ErrConnectionRequired, got %v", err)
	}
}

// staticIdentityDefinition builds a definition whose resolver always answers with instance id
func staticIdentityDefinition(instanceID string) types.Definition {
	return types.Definition{
		Installation: &types.InstallationRegistration{
			Resolve: func(context.Context, types.InstallationRequest) (types.IntegrationInstallationMetadata, bool, error) {
				return types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: instanceID}}, true, nil
			},
		},
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

func TestResolveConnectionIdentityDoesNotRejectInstanceMismatch(t *testing.T) {
	t.Parallel()

	installation := &ent.Integration{
		InstallationMetadata: types.IntegrationInstallationMetadata{Display: types.IntegrationInstallationIdentity{ExternalID: "tenant-a"}},
	}

	metadata, err := resolveConnectionIdentity(context.Background(), installation, staticIdentityDefinition("tenant-b"), types.ConnectionRegistration{}, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if metadata.Display.ExternalID != "tenant-b" {
		t.Fatalf("expected tenant-b, got %q", metadata.Display.ExternalID)
	}
}

func TestResolveConnectionIdentityRejectsResolverWithoutInstanceID(t *testing.T) {
	t.Parallel()

	def := types.Definition{
		Installation: &types.InstallationRegistration{
			Resolve: func(context.Context, types.InstallationRequest) (types.IntegrationInstallationMetadata, bool, error) {
				return types.IntegrationInstallationMetadata{}, false, nil
			},
		},
	}

	_, err := resolveConnectionIdentity(context.Background(), &ent.Integration{}, def, types.ConnectionRegistration{}, nil, nil)
	if !errors.Is(err, ErrInstallationInstanceIDRequired) {
		t.Fatalf("expected ErrInstallationInstanceIDRequired, got %v", err)
	}
}

func TestResolveConnectionIdentityFallsBackToSelfWithoutInstallation(t *testing.T) {
	t.Parallel()

	installation := &ent.Integration{ID: "installation-id"}

	metadata, err := resolveConnectionIdentity(context.Background(), installation, types.Definition{}, types.ConnectionRegistration{}, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if metadata.Display.ExternalID != installation.ID {
		t.Fatalf("expected the installation's own id, got %q", metadata.Display.ExternalID)
	}
}
