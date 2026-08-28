package runtime

import (
	"testing"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestIntegrationUnhealthyReason(t *testing.T) {
	t.Parallel()

	if got := IntegrationUnhealthyReason(&ent.Integration{}); got != "" {
		t.Fatalf("IntegrationUnhealthyReason = %q, want empty for healthy installation", got)
	}

	installation := &ent.Integration{Metadata: map[string]any{
		unhealthyReasonMetadataKey: "the connection needs to be reauthorized",
	}}

	if got := IntegrationUnhealthyReason(installation); got != "the connection needs to be reauthorized" {
		t.Fatalf("IntegrationUnhealthyReason = %q, want recorded reason", got)
	}
}

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
