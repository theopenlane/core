package catalog

import (
	"os"
	"path/filepath"
	"testing"

	yaml "github.com/goccy/go-yaml"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestComputeSHAAndSaveCatalog(t *testing.T) {
	c := New()

	c.Version = "v0.0.1"
	c.Modules = FeatureSet{
		"m1": {DisplayName: "M1", Billing: Billing{Prices: []Price{{Interval: "month", UnitAmount: 100}}}},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.yaml")

	diff, err := c.SaveCatalog(path)
	assert.Check(t, err)
	assert.Check(t, "" != diff)

	data, err := os.ReadFile(path)
	assert.Check(t, err)

	var saved Catalog
	assert.Check(t, yaml.Unmarshal(data, &saved))

	assert.Check(t, is.Equal(c.Version, saved.Version))
	assert.Check(t, is.Equal(computeSHA(saved.Version), saved.SHA))
}

func TestSaveCatalogNoChanges(t *testing.T) {
	c := New()
	c.Version = "v0.0.1"

	c.Modules = FeatureSet{
		"m1": {DisplayName: "M1", Billing: Billing{Prices: []Price{{Interval: "month", UnitAmount: 100}}}},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.yaml")

	_, err := c.SaveCatalog(path)
	assert.Check(t, err)

	firstVersion := c.Version
	firstSHA := c.SHA

	diff, err := c.SaveCatalog(path)
	assert.Check(t, err)
	assert.Check(t, is.Equal("", diff))
	assert.Check(t, is.Equal(firstVersion, c.Version))
	assert.Check(t, is.Equal(firstSHA, c.SHA))
}
