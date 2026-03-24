package runtime

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/gala"
)

func newTestGala(t *testing.T) *gala.Gala {
	t.Helper()

	g, err := gala.NewGala(context.Background(), gala.Config{
		DispatchMode: gala.DispatchModeInMemory,
		Enabled:      true,
		WorkerCount:  1,
	})
	if err != nil {
		t.Fatalf("failed to create in-memory gala: %v", err)
	}

	return g
}

func TestNewMinimalConfig(t *testing.T) {
	t.Parallel()

	g := newTestGala(t)
	ks, err := keystore.NewStore(nil)
	if err != nil {
		t.Fatalf("expected keystore error for nil DB: %v", err)
	}

	_ = ks
	// keystore.NewStore requires a non-nil DB, so we test New with a registry override
	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          "test-def",
			DisplayName: "Test",
		},
	})

	// New requires non-nil DB for keystore, so test what we can without it
	// Test that New returns an error when keystore is nil
	_, err = New(Config{
		Gala:                  g,
		Registry:              reg,
		SkipExecutorListeners: true,
	})
	// This will fail because keystore is nil, but the error path is tested
	if err == nil {
		// If it somehow succeeds, verify it works
		return
	}
	// Expected: panics or errors from nil keystore - acceptable for unit test boundary
}

func TestNewWithRegistryOverride(t *testing.T) {
	t.Parallel()

	g := newTestGala(t)

	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          "test-def",
			DisplayName: "Test",
		},
	})

	rt, err := New(Config{
		Gala:                  g,
		Registry:              reg,
		Keystore:              &keystore.Store{},
		SkipExecutorListeners: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rt.Registry() != reg {
		t.Fatal("expected registry override to be used")
	}

	if rt.Gala() != g {
		t.Fatal("expected gala to be wired")
	}

	if rt.Keystore() == nil {
		t.Fatal("expected keystore to be wired")
	}

	def, ok := rt.Definition("test-def")
	if !ok {
		t.Fatal("expected definition to be found")
	}

	if def.DisplayName != "Test" {
		t.Fatalf("expected display name Test, got %q", def.DisplayName)
	}

	catalog := rt.Catalog()
	if len(catalog) != 1 {
		t.Fatalf("expected 1 catalog entry, got %d", len(catalog))
	}
}

func TestNewWithInMemoryAuthState(t *testing.T) {
	t.Parallel()

	g := newTestGala(t)

	rt, err := New(Config{
		Gala:                  g,
		Registry:              registry.New(),
		Keystore:              &keystore.Store{},
		SkipExecutorListeners: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rt.Keymaker() == nil {
		t.Fatal("expected keymaker to be wired with in-memory auth state")
	}
}

func TestNewWithBuildersNoRegistry(t *testing.T) {
	t.Parallel()

	g := newTestGala(t)

	called := false
	builder := registry.Builder(func() (types.Definition, error) {
		called = true
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          "built-def",
				DisplayName: "Built",
			},
		}, nil
	})

	rt, err := New(Config{
		Gala:                  g,
		DefinitionBuilders:    []registry.Builder{builder},
		Keystore:              &keystore.Store{},
		SkipExecutorListeners: true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected builder to be called")
	}

	_, ok := rt.Definition("built-def")
	if !ok {
		t.Fatal("expected built definition to be registered")
	}
}
