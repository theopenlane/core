package catalog_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
)

func writeTempCatalog(t *testing.T, data string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "catalog.yaml")
	assert.NoError(t, os.WriteFile(p, []byte(data), 0o644))
	return p
}

func TestLoadCatalog(t *testing.T) {
	t.Parallel()
	yaml := `version: v0.0.1
sha: ae5bcf31543244e0bc0b0a14a4374ae2f199eebe805de0ad58f20d36b5d5649b
modules:
  mod1:
    display_name: M1
    lookup_key: mod1
    description: This is a module
    billing:
      prices:
        - interval: month
          unit_amount: 100
          nickname: mod1_monthly
          lookup_key: mod1_monthly
    audience: public
addons:
  add1:
    display_name: A1
    lookup_key: add1
    description: Addon description
    billing:
      prices:
        - interval: month
          unit_amount: 50
    audience: private
`
	path := writeTempCatalog(t, yaml)

	c, err := catalog.LoadCatalog(path)
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Contains(t, c.Modules, "mod1")
	assert.Contains(t, c.Addons, "add1")
}

func TestLoadCatalogMissing(t *testing.T) {
	t.Parallel()
	_, err := catalog.LoadCatalog("/no/such/file")
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestFreeModules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		modules  catalog.FeatureSet
		expected []models.OrgModule
	}{
		{
			name: "returns only modules where all prices are zero",
			modules: catalog.FeatureSet{
				"free_mod": {Billing: catalog.Billing{Prices: []catalog.Price{
					{Interval: "month", UnitAmount: 0},
					{Interval: "year", UnitAmount: 0},
				}}},
				"paid_mod": {Billing: catalog.Billing{Prices: []catalog.Price{
					{Interval: "month", UnitAmount: 1000},
				}}},
			},
			expected: []models.OrgModule{"free_mod"},
		},
		{
			name: "module with mixed prices is not free",
			modules: catalog.FeatureSet{
				"mixed_mod": {Billing: catalog.Billing{Prices: []catalog.Price{
					{Interval: "month", UnitAmount: 0},
					{Interval: "year", UnitAmount: 500},
				}}},
			},
			expected: nil,
		},
		{
			name:     "no modules returns empty",
			modules:  catalog.FeatureSet{},
			expected: nil,
		},
		{
			name: "all free modules returned",
			modules: catalog.FeatureSet{
				"free_a": {Billing: catalog.Billing{Prices: []catalog.Price{{Interval: "month", UnitAmount: 0}}}},
				"free_b": {Billing: catalog.Billing{Prices: []catalog.Price{{Interval: "month", UnitAmount: 0}}}},
			},
			expected: []models.OrgModule{"free_a", "free_b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := catalog.New()
			c.Modules = tt.modules

			got := c.FreeModules()

			slices.Sort(got)
			slices.Sort(tt.expected)

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFreeModulesDefaultCatalog(t *testing.T) {
	t.Parallel()

	c := catalog.Catalog{Catalog: gencatalog.DefaultCatalog}
	free := c.FreeModules()

	assert.Contains(t, free, models.CatalogBaseModule)
	assert.Contains(t, free, models.CatalogRegistryModule)
}

func TestVisible(t *testing.T) {
	t.Parallel()
	c := catalog.New()
	c.Modules = catalog.FeatureSet{
		"m1": {Audience: "public"},
		"m2": {Audience: "beta"},
	}
	c.Addons = catalog.FeatureSet{
		"a1": {Audience: "private"},
		"a2": {Audience: "public"},
	}

	pub := c.Visible("public")
	assert.Len(t, pub.Modules, 1)
	assert.NotContains(t, pub.Modules, "m2")
	assert.Len(t, pub.Addons, 1)
	assert.Contains(t, pub.Addons, "a2")

	all := c.Visible("")
	assert.Len(t, all.Modules, 2)
	assert.Len(t, all.Addons, 2)
}
