package hooks

import (
	"testing"

	"github.com/theopenlane/core/pkg/gala"
)

func TestResolveGalaRuntimes(t *testing.T) {
	runtimeA, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create first runtime: %v", err)
	}

	runtimeB, err := gala.NewInMemory()
	if err != nil {
		t.Fatalf("failed to create second runtime: %v", err)
	}

	providers := []func() *gala.Gala{
		nil,
		func() *gala.Gala { return runtimeA },
		func() *gala.Gala { return nil },
		func() *gala.Gala { return runtimeA },
		func() *gala.Gala { return runtimeB },
	}

	runtimes := resolveGalaRuntimes(providers)
	if len(runtimes) != 2 {
		t.Fatalf("expected two runtimes, got %d", len(runtimes))
	}

	if runtimes[0] != runtimeA {
		t.Fatalf("expected first runtime to match runtimeA")
	}

	if runtimes[1] != runtimeB {
		t.Fatalf("expected second runtime to match runtimeB")
	}
}
