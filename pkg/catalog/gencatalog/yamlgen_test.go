package gencatalog_test

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/pkg/catalog/gencatalog"
)

// TestDefaultCatalogTypedModules ensures the default catalog contains
// entries for every defined OrgModule constant
func TestDefaultCatalogTypedModules(t *testing.T) {
	all := map[string]struct{}{}
	for k := range gencatalog.DefaultCatalog.Modules {
		all[k] = struct{}{}
	}

	for k := range gencatalog.DefaultCatalog.Addons {
		all[k] = struct{}{}
	}

	for _, mod := range models.AllOrgModules {
		_, ok := all[mod.String()]
		assert.Check(t, ok, "missing module %s", mod)
	}

	assert.Check(t, is.Len(all, len(models.AllOrgModules)))
}
