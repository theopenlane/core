package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerProductCatalogHandler registers the /products handler
func registerProductCatalogHandler(router *Router) error {
	// add route without the path param
	config := Config{
		Path:        "/products",
		Method:      http.MethodGet,
		Name:        "ProductCatalog",
		Description: "List products available in the product catalog",
		Tags:        []string{"catalog"},
		OperationID: "ProductCatalog",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		Handler:     router.Handler.ProductCatalogHandler,
	}

	return router.AddV1HandlerRoute(config)
}
