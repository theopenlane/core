package operations

import (
	"errors"
	"testing"
)

func TestRegisterRuntimeListeners_NilRuntime(t *testing.T) {
	t.Parallel()

	err := RegisterRuntimeListeners(nil, nil, nil, nil)
	if !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}
