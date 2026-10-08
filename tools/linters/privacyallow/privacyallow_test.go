package privacyallow_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/tools/linters/privacyallow"
)

func TestPrivacyAllow(t *testing.T) {
	plugin, err := privacyallow.New(map[string]any{})
	assert.NilError(t, err)

	analyzers, err := plugin.BuildAnalyzers()
	assert.NilError(t, err)

	analysistest.Run(t, analysistest.TestData(), analyzers[0], "internal/integrations/listener", "outofscope", "internal/ent/generated/authz")
}
