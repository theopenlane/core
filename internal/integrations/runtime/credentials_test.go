package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// --- resolveConnectionFromState ---

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

	credRef := types.NewCredentialSlotID("oauth")
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credRef, Name: "OAuth", Schema: json.RawMessage(`{"type":"object"}`)},
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

// --- resolveConnectionForCredential ---

func TestResolveConnectionForCredentialFromPersistedState(t *testing.T) {
	t.Parallel()

	credRef := types.NewCredentialSlotID("oauth")
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credRef, Name: "OAuth", Schema: json.RawMessage(`{"type":"object"}`)},
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

	credRef := types.NewCredentialSlotID("oauth")
	otherRef := types.NewCredentialSlotID("api-key")
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credRef, Name: "OAuth", Schema: json.RawMessage(`{"type":"object"}`)},
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

	credRef := types.NewCredentialSlotID("api-key")
	def := types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credRef, Name: "API Key", Schema: json.RawMessage(`{"type":"object"}`)},
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

// --- Disconnect ---

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

func TestDisconnectConnectionResolutionError(t *testing.T) {
	t.Parallel()

	credRef := types.NewCredentialSlotID("unknown")
	providerState, _ := json.Marshal(types.DefinitionProviderState{CredentialRef: credRef})

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
	})

	rt := NewForTesting(reg)
	_, err := rt.Disconnect(context.Background(), &ent.Integration{
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

// --- Reconcile ---

func TestReconcileNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.Reconcile(context.Background(), nil, nil, types.CredentialSlotID{}, nil, nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestReconcileMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	err := rt.Reconcile(context.Background(), &ent.Integration{
		DefinitionID: "nonexistent",
	}, nil, types.CredentialSlotID{}, nil, nil)
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
		DefinitionID: "test-def",
	}, nil, types.CredentialSlotID{}, nil, nil)
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
	// Empty JSON objects and null are treated as empty by jsonx.IsEmptyRawMessage
	err := rt.Reconcile(context.Background(), &ent.Integration{
		DefinitionID: "test-def",
	}, json.RawMessage(`null`), types.CredentialSlotID{}, nil, nil)
	if err != nil {
		t.Fatalf("expected no error for null input reconcile, got %v", err)
	}
}

// --- resolvePersistedConnection (additional cases beyond definitions_test.go) ---

func TestResolvePersistedConnectionSingleConnectionFallback(t *testing.T) {
	t.Parallel()

	credRef := types.NewCredentialSlotID("oauth")
	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def"},
		CredentialRegistrations: []types.CredentialRegistration{
			{Ref: credRef, Name: "OAuth", Schema: json.RawMessage(`{"type":"object"}`)},
		},
		Connections: []types.ConnectionRegistration{
			{CredentialRef: credRef, Name: "Only Connection", CredentialRefs: []types.CredentialSlotID{credRef}},
		},
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

	refA := types.NewCredentialSlotID("oauth")
	refB := types.NewCredentialSlotID("api-key")

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
