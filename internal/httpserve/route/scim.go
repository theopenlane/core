package route

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers/scim"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
)

const (
	// scimRoutePrefix is the URL prefix stripped before dispatching to the SCIM library
	scimRoutePrefix = "/v1/integrations/scim"
)

// registerSCIMRoutes sets up SCIM routes on the public endpoint using per-installation
// Bearer token authentication resolved through the integration webhook infrastructure
func registerSCIMRoutes(router *Router) error {
	server, err := scim.NewSCIMServer()
	if err != nil {
		return err
	}

	scimHandler := scim.WrapSCIMServerHTTPHandler(server)

	handler := func(c echo.Context) error {
		ctx := c.Request().Context()

		endpointID := c.PathParam("endpointID")
		if endpointID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing endpoint ID")
		}

		bearerToken := extractBearerToken(c.Request())
		if bearerToken == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
		}

		// Bypass privacy rules for webhook-style resolution
		ctx = privacy.DecisionContext(ctx, privacy.Allow)
		ctx = auth.WithCaller(ctx, auth.NewWebhookCaller(""))

		rt := router.Handler.IntegrationsRuntime
		webhook, err := rt.ResolveWebhookByEndpoint(ctx, endpointID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "unknown SCIM endpoint")
		}

		if subtle.ConstantTimeCompare([]byte(bearerToken), []byte(webhook.SecretToken)) != 1 {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}

		installation, err := rt.ResolveIntegration(ctx, "", webhook.IntegrationID, "")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "integration not found")
		}

		if installation.DefinitionID != definitionscim.DefinitionID.ID() {
			return echo.NewHTTPError(http.StatusForbidden, "integration is not a SCIM installation")
		}

		if installation.Status != enums.IntegrationStatusConnected {
			return echo.NewHTTPError(http.StatusForbidden, "integration is not active")
		}

		// Narrow caller to the installation owner's org
		ctx = auth.WithCaller(ctx, auth.NewWebhookCaller(installation.OwnerID))

		ctx = scim.WithSCIMRequest(ctx, &scim.SCIMRequest{
			Installation: installation,
			Runtime:      rt,
			BasePath:     fmt.Sprintf("%s/%s/v2", scimRoutePrefix, endpointID),
		})

		// Strip URL prefix so the SCIM library sees /v2/...
		req := c.Request()
		originalPath := req.URL.Path
		trimPrefix := fmt.Sprintf("%s/%s", scimRoutePrefix, endpointID)
		newURL := *req.URL
		newURL.Path = strings.TrimPrefix(originalPath, trimPrefix)

		reqWithPath := req.Clone(ctx)
		reqWithPath.URL = &newURL

		scimHandler(c.Response(), reqWithPath)

		return nil
	}

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

// extractBearerToken parses the Bearer token from the Authorization header
func extractBearerToken(r *http.Request) string {
	const bearerPrefix = "Bearer "

	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(header, bearerPrefix)
}
