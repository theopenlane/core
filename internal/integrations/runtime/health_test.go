package runtime

import (
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestIntegrationDisplayName(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	if err := reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{ID: "test-def", DisplayName: "Test Provider"},
	}); err != nil {
		t.Fatalf("registering definition: %v", err)
	}

	rt := NewForTesting(reg)

	if got := rt.integrationDisplayName(&ent.Integration{DefinitionID: "test-def", Name: "my install"}); got != "Test Provider" {
		t.Fatalf("integrationDisplayName = %q, want definition display name", got)
	}

	if got := rt.integrationDisplayName(&ent.Integration{DefinitionID: "gone-def", Name: "my install"}); got != "my install" {
		t.Fatalf("integrationDisplayName = %q, want installation name fallback", got)
	}
}
