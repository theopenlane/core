//go:build cli

package speccli

import "testing"

func TestRegisterOverride(t *testing.T) {
	const key = "example"

	RegisterOverride(key, func(spec *CommandSpec) error {
		spec.Short = "overridden"
		return nil
	})

	override, ok := lookupOverride(key)
	if !ok {
		t.Fatalf("expected override to be registered")
	}

	spec := CommandSpec{Short: "original"}
	if err := override(&spec); err != nil {
		t.Fatalf("unexpected error applying override: %v", err)
	}
	if spec.Short != "overridden" {
		t.Fatalf("expected override to mutate spec, got %q", spec.Short)
	}

	// Remove override and verify it is gone.
	RegisterOverride(key, nil)
	if _, ok := lookupOverride(key); ok {
		t.Fatalf("expected override to be removed")
	}
}
