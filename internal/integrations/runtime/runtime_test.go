package runtime

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/operations"
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

	// keystore.NewStore requires a non-nil DB and returns an error for nil
	_, err := keystore.NewStore(nil)
	if err == nil {
		t.Fatal("expected error from keystore.NewStore(nil)")
	}

	// New without a keystore should panic or error at wiring time
	g := newTestGala(t)
	reg := registry.New()
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID:          "test-def",
			DisplayName: "Test",
		},
	})

	_, err = New(Config{
		Gala:     g,
		Registry: reg,
	})
	if err == nil {
		return
	}
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
		Gala:     g,
		Registry: reg,
		Keystore: &keystore.Store{},
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

	if rt.keystore() == nil {
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
		Gala:     g,
		Registry: registry.New(),
		Keystore: &keystore.Store{},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rt.keymaker() == nil {
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
		Gala:               g,
		DefinitionBuilders: []registry.Builder{builder},
		Keystore:           &keystore.Store{},
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

func TestNewDefaultLookback(t *testing.T) {
	t.Parallel()

	g := newTestGala(t)
	rt, err := New(Config{
		Gala:     g,
		Registry: registry.New(),
		Keystore: &keystore.Store{},
	})
	assert.NilError(t, err)
	assert.Equal(t, rt.defaultLookback, defaultLookbackDuration)
}

func TestNewCustomLookback(t *testing.T) {
	t.Parallel()

	g := newTestGala(t)
	custom := 30 * 24 * time.Hour

	rt, err := New(Config{
		Gala:            g,
		Registry:        registry.New(),
		Keystore:        &keystore.Store{},
		DefaultLookback: custom,
	})
	assert.NilError(t, err)
	assert.Equal(t, rt.defaultLookback, custom)
}

func TestNormalizeDispatchError(t *testing.T) {
	t.Parallel()

	unexpectedErr := errors.New("boom")

	tests := []struct {
		name        string
		err         error
		expectedErr error
	}{
		{
			name:        "nil",
			err:         nil,
			expectedErr: nil,
		},
		{
			name:        "definition not found",
			err:         registry.ErrDefinitionNotFound,
			expectedErr: ErrDefinitionNotFound,
		},
		{
			name:        "wrapped operation not found",
			err:         fmt.Errorf("wrapped: %w", registry.ErrOperationNotFound),
			expectedErr: ErrOperationNotFound,
		},
		{
			name:        "dispatch input invalid",
			err:         operations.ErrDispatchInputInvalid,
			expectedErr: operations.ErrDispatchInputInvalid,
		},
		{
			name:        "unexpected error",
			err:         unexpectedErr,
			expectedErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeDispatchError(tt.err)
			switch {
			case tt.expectedErr == nil:
				if got != nil {
					t.Fatalf("expected nil error, got %v", got)
				}
			case !errors.Is(got, tt.expectedErr):
				t.Fatalf("expected %v, got %v", tt.expectedErr, got)
			}
		})
	}
}
