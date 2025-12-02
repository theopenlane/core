package gencatalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/catalog/gencatalog"
	"github.com/theopenlane/shared/models"
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
		assert.True(t, ok, "missing module %s", mod)
	}

	assert.Len(t, all, len(models.AllOrgModules))
}
