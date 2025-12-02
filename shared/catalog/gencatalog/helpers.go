package gencatalog

import (
	catalog "github.com/theopenlane/shared/catalog"
)

// GetDefaultCatalog returns the default catalog; if useSandbox is true, it returns the sandbox catalog
func GetDefaultCatalog(useSandbox bool) catalog.Catalog {
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

	return c.Visible(audience)
}
