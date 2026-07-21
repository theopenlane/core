package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerCloudflareSnapshotHandler registers the Cloudflare snapshot handler and route
func registerCloudflareSnapshotHandler(router *Router) error {
	config := Config{
		Path:         "/snapshot",
		Method:       http.MethodPost,
		Name:         "CloudflareSnapshot",
		Description:  "Take a snapshot using Cloudflare",
		Tags:         []string{"Scans"},
		OperationID:  "CloudflareSnapshot",
		Security:     handlers.AuthenticatedSecurity,
		IncludeInOAS: true,
		Middlewares:  *authenticatedEndpoint,
		Handler:      router.Handler.SnapshotHandler,
	}

	return router.AddV1HandlerRoute(config)
}
