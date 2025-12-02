package catalog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/catalog"
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

func TestVisible(t *testing.T) {
	t.Parallel()
	c := &catalog.Catalog{
		Modules: catalog.FeatureSet{
			"m1": {Audience: "public"},
			"m2": {Audience: "beta"},
		},
		Addons: catalog.FeatureSet{
			"a1": {Audience: "private"},
			"a2": {Audience: "public"},
		},
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
