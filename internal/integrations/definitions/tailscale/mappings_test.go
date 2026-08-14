package tailscale

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMembershipSnapshotCompleteness(t *testing.T) {
	assert.Equal(t, membershipSnapshotCompleteness(nil), types.SnapshotCompletenessFull)
	assert.Equal(t, membershipSnapshotCompleteness(errors.New("policy unavailable")), types.SnapshotCompletenessPartial)
}

func TestMappingExpressionsValid(t *testing.T) {
	def, err := Builder()()
	assert.NilError(t, err)

	for _, m := range def.Mappings {
		name := m.Schema
		if m.Variant != "" {
			name += "/" + m.Variant
		}

		t.Run(name+"/filter", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.FilterExpr))
		})

		t.Run(name+"/map", func(t *testing.T) {
			assert.NilError(t, providerkit.ValidateExpr(m.Spec.MapExpr))
		})
	}
}
