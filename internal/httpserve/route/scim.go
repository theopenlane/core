package route

import (
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/httpserve/handlers/scim"
)

// registerSCIMRoutes sets up SCIM routes
func registerSCIMRoutes(router *Router) error {
	server, err := scim.NewSCIMServer()
	if err != nil {
		return err
	}

	scimHandler := scim.WrapSCIMServerHTTPHandler(server)

	handler := func(c echo.Context) error {
		// Get the request with updated context (after middleware)
		req := c.Request()

		// Strip /scim prefix so the library sees the path it expects
		originalPath := req.URL.Path
		newURL := *req.URL
		newURL.Path = strings.TrimPrefix(originalPath, "/scim")

		// Create new request with updated URL path and preserved context
		reqWithPath := req.Clone(req.Context())
		reqWithPath.URL = &newURL

		scimHandler(c.Response(), reqWithPath)

		return nil
	}

	grp := router.Base()

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	grp.Match(methods, "/scim/*", handler, *authenticatedEndpoint...)

	return nil
}
