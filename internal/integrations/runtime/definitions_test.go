package runtime

import (
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
			ID: "test-def",
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
