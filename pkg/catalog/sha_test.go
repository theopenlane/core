package catalog

import (
	"os"
	"path/filepath"
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeSHAAndSaveCatalog(t *testing.T) {
	c := &Catalog{
		Version: "v0.0.1",
		Modules: FeatureSet{
			"m1": {DisplayName: "M1", Billing: Billing{Prices: []Price{{Interval: "month", UnitAmount: 100}}}},
		},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.yaml")

	diff, err := c.SaveCatalog(path)
	require.NoError(t, err)
	assert.NotEqual(t, "", diff)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var saved Catalog
	require.NoError(t, yaml.Unmarshal(data, &saved))

	assert.Equal(t, c.Version, saved.Version)
	assert.Equal(t, computeSHA(saved.Version), saved.SHA)
}

func TestSaveCatalogNoChanges(t *testing.T) {
	c := &Catalog{
		Version: "v0.0.1",
		Modules: FeatureSet{
			"m1": {DisplayName: "M1", Billing: Billing{Prices: []Price{{Interval: "month", UnitAmount: 100}}}},
		},
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.yaml")

	_, err := c.SaveCatalog(path)
	require.NoError(t, err)

	firstVersion := c.Version
	firstSHA := c.SHA

	diff, err := c.SaveCatalog(path)
	require.NoError(t, err)
	assert.Equal(t, "", diff)
	assert.Equal(t, firstVersion, c.Version)
	assert.Equal(t, firstSHA, c.SHA)
}
