package gencatalog

import (
	models "github.com/theopenlane/core/common/models"
	catalog "github.com/theopenlane/core/pkg/catalog"
)

// GetDefaultCatalog returns the default catalog; if useSandbox is true, it returns the sandbox catalog
func GetDefaultCatalog(useSandbox bool) models.Catalog {
	if useSandbox {
		return DefaultSandboxCatalog
	}

	return DefaultCatalog
}

// GetModules returns a list of all module names in the default catalog
func GetModules(useSandbox bool) catalog.FeatureSet {
	c := GetDefaultCatalog(useSandbox)

	return c.Modules
}

// GetCatalogByAudience returns a catalog filtered by audience; if useSandbox is true, it uses the sandbox catalog
func GetCatalogByAudience(useSandbox bool, audience string) *catalog.Catalog {
	c := GetDefaultCatalog(useSandbox)

	cat := catalog.Catalog{
		Catalog: c,
	}

	return cat.Visible(audience)
}
