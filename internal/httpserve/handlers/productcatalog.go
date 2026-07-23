package handlers

import (
	commonmodels "github.com/theopenlane/core/common/models"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/catalog"
	"github.com/theopenlane/core/pkg/catalog/gencatalog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"
)

// ProductCatalogHandler lists all products in the catalog
func (h *Handler) ProductCatalogHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.ProductCatalogRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	out := &models.ProductCatalogResponse{
		Reply:   rout.Reply{Success: true},
		Catalog: h.filterCatalog(in),
	}

	return h.Success(ctx, out)
}

// filterCatalog filters the catalog based on the request parameters
func (h *Handler) filterCatalog(in *models.ProductCatalogRequest) commonmodels.Catalog {
	return catalog.FilterByAudience(in.IncludeBeta, in.IncludePrivate, gencatalog.GetDefaultCatalog(h.DBClient.EntConfig.Modules.UseSandbox))
}
