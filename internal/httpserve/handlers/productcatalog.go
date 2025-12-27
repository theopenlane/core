package handlers

import (
	commonmodels "github.com/theopenlane/core/common/models"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"
)

const (
	// PrivateAudience defines a product only not publicly available
	PrivateAudience = "private"
	// BetaAudience defines a product available to beta users
	BetaAudience = "beta"
	// PublicAudience defines a product available to all users
	PublicAudience = "public"
)

// ProductCatalogHandler lists all products in the catalog
func (h *Handler) ProductCatalogHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, models.ExampleProductCatalogRequest, models.ExampleProductCatalogReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	out := &models.ProductCatalogReply{
		Reply:   rout.Reply{Success: true},
		Catalog: h.filterCatalog(in),
	}

	return h.Success(ctx, out, openapi)
}

// filterCatalog filters the catalog based on the request parameters
func (h *Handler) filterCatalog(in *models.ProductCatalogRequest) commonmodels.Catalog {
	cat := gencatalog.DefaultCatalog

	if h.DBClient.EntConfig.Modules.UseSandbox {
		cat = gencatalog.DefaultSandboxCatalog
	}

	if in.IncludeBeta && in.IncludePrivate {
		return cat
	}

	return doFilter(in.IncludeBeta, in.IncludePrivate, cat)
}

// doFilter filters loops through the modules and addons and filters them based on the audience
func doFilter(includeBeta, includePrivate bool, cat commonmodels.Catalog) commonmodels.Catalog {
	filtered := cat

	modules := catalog.FeatureSet{}

	for k, v := range cat.Modules {
		if include(v.Audience, includeBeta, includePrivate) {
			modules[k] = v
		}
	}

	addons := catalog.FeatureSet{}

	for k, v := range cat.Addons {
		if include(v.Audience, includeBeta, includePrivate) {
			addons[k] = v
		}
	}

	filtered.Modules = modules
	filtered.Addons = addons

	return filtered
}

// include returns true if the audience should be included based on the request parameters
func include(aud string, includeBeta, includePrivate bool) bool {
	switch aud {
	case PublicAudience:
		return true
	case BetaAudience:
		return includeBeta
	case PrivateAudience:
		return includePrivate
	default:
		return false
	}
}
