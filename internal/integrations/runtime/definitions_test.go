package runtime

import (
	"encoding/json"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestResolveDefinitionForInstallationNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.resolveDefinitionForInstallation(nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestResolveDefinitionForInstallationMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.resolveDefinitionForInstallation(&ent.Integration{
		DefinitionID: "nonexistent",
	})
	if !errors.Is(err, registry.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestResolveDefinitionForInstallationSuccess(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:   "test-def",
		},
	})

	rt := NewForTesting(reg)
	def, err := rt.resolveDefinitionForInstallation(&ent.Integration{
		DefinitionID: "test-def",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if def.ID != "test-def" {
		t.Fatalf("expected ID test-def, got %q", def.ID)
	}
}

func TestResolvePersistedConnectionNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.ResolvePersistedConnection(nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestResolvePersistedConnectionMissingDefinition(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.ResolvePersistedConnection(&ent.Integration{
		DefinitionID: "nonexistent",
	})
	if !errors.Is(err, registry.ErrDefinitionNotFound) {
		t.Fatalf("expected ErrDefinitionNotFound, got %v", err)
	}
}

func TestResolvePersistedConnectionNoConnections(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:   "test-def",
		},
	})

	rt := NewForTesting(reg)
	_, err := rt.ResolvePersistedConnection(&ent.Integration{
		DefinitionID: "test-def",
	})
	if !errors.Is(err, ErrConnectionRequired) {
		t.Fatalf("expected ErrConnectionRequired, got %v", err)
	}
}

func TestResolvePersistedConnectionMatchesPersistedRef(t *testing.T) {
	t.Parallel()

	credRef := types.NewCredentialSlotID("oauth")
	reg := registry.New()
	err := reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:   "test-def",
		},
		CredentialRegistrations: []types.CredentialRegistration{
			{
				Ref:    credRef,
				Name:   "OAuth Credential",
				Schema: json.RawMessage(`{"type":"object"}`),
			},
		},
		Connections: []types.ConnectionRegistration{
			{
				CredentialRef:  credRef,
				Name:           "OAuth Connection",
				CredentialRefs: []types.CredentialSlotID{credRef},
			},
		},
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	providerState, err := json.Marshal(types.DefinitionProviderState{
		CredentialRef: credRef,
	})
	if err != nil {
		t.Fatalf("marshal provider state: %v", err)
	}

	rt := NewForTesting(reg)
	conn, err := rt.ResolvePersistedConnection(&ent.Integration{
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

	if conn.Name != "OAuth Connection" {
		t.Fatalf("expected connection name OAuth Connection, got %q", conn.Name)
	}
}

func TestNewForTestingRegistry(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID: "def-1",
		},
	})

	rt := NewForTesting(reg)
	if rt.Registry() != reg {
		t.Fatal("expected registry to match")
	}

	_, ok := rt.Definition("def-1")
	if !ok {
		t.Fatal("expected definition to be found")
	}

	catalog := rt.Catalog()
	if len(catalog) != 1 {
		t.Fatalf("expected 1 catalog entry, got %d", len(catalog))
	}
}
