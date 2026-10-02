package runtime

import (
	"encoding/json"
	"testing"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestWorkloadOperations(t *testing.T) {
	t.Parallel()

	def := types.Definition{Operations: []types.OperationRegistration{
		{Name: "sync"},
		{Name: "internal", Internal: true},
		{Name: "disabled-globally", DisabledForAll: true},
		{Name: "disabled-for-install"},
		{Name: "enabled-for-install"},
	}}

	installation := &ent.Integration{OperationConfig: openapi.IntegrationOperationConfig{Operations: map[string]json.RawMessage{
		"disabled-for-install": json.RawMessage(`{"disable":true}`),
		"enabled-for-install":  json.RawMessage(`{"disable":false}`),
	}}}

	got := lo.Map(workloadOperations(def, installation), func(op types.OperationRegistration, _ int) string {
		return op.Name
	})

	want := []string{"sync", "enabled-for-install"}
	if len(got) != len(want) {
		t.Fatalf("workloadOperations = %v, want %v", got, want)
	}

	for i, name := range want {
		if got[i] != name {
			t.Fatalf("workloadOperations = %v, want %v", got, want)
		}
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
