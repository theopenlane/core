package operations

import (
	"errors"
	"testing"

	"github.com/theopenlane/core/pkg/gala"
)

func TestRegisterRuntimeListeners_NilRuntime(t *testing.T) {
	t.Parallel()

	err := RegisterRuntimeListeners(nil, nil, nil, nil, nil, gala.Schedule{}, nil)
	if !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}
