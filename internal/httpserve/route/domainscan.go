package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

// registerDomainScanHandler registers the domain scan handler
func registerDomainScanHandler(router *Router) error {
	config := Config{
		Path:        "/domain-scan",
		Method:      http.MethodPost,
		Name:        "DomainScan",
		Description: "Queue a domain scan for a single domain. The scan runs asynchronously and returns the created scan id, gated to one request per organization per hour",
		Tags:        []string{"domainscan"},
		OperationID: "DomainScan",
		Security:    handlers.AuthenticatedSecurity,
		Middlewares: *authenticatedEndpoint,
		Handler:     router.Handler.DomainScanHandler,
	}

	return router.AddV1HandlerRoute(config)
}
