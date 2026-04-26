package okta

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

func TestMappingExpressionsValid(t *testing.T) {
	for _, m := range oktaMappings() {
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
