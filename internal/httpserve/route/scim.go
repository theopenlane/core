package route

import (
	"net/http"

	"github.com/theopenlane/core/internal/httpserve/handlers/scim"
)

const (
	// scimRoutePrefix is the URL prefix stripped before dispatching to the SCIM library
	scimRoutePrefix = "/v1/integrations/scim"
)

// registerSCIMRoutes sets up SCIM routes on the public endpoint using per-installation
// Bearer token authentication resolved through the integration webhook infrastructure
func registerSCIMRoutes(router *Router) error {
	server, err := scim.NewSCIMServer(router.Handler.IntegrationsRuntime)
	if err != nil {
		return err
	}

	handler := router.Handler.SCIMHandler(server, scimRoutePrefix)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	grp := router.VersionOne()
	grp.Match(methods, "/integrations/scim/:endpointID/*", handler, *unauthenticatedEndpoint...)

	return nil
}
