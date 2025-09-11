package handlers

import (
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
	models "github.com/theopenlane/core/pkg/openapi"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"
)

const (
	PrivateAudience = "private"
	BetaAudience    = "beta"
	PublicAudience  = "public"
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

func (h *Handler) filterCatalog(in *models.ProductCatalogRequest) catalog.Catalog {
	cat := gencatalog.DefaultCatalog

	if h.DBClient.EntConfig.Modules.UseSandbox {
		cat = gencatalog.DefaultSandboxCatalog
	}

	if in.IncludeBeta && in.IncludePrivate {
		return cat
	}

	return doFilter(in.IncludeBeta, in.IncludePrivate, cat)
}

func doFilter(includeBeta, includePrivate bool, cat catalog.Catalog) catalog.Catalog {
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
