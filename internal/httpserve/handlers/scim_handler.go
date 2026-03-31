package handlers

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
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

// SCIMHandler resolves the SCIM endpoint, authenticates the bearer token,
// validates the integration installation, and dispatches through the SCIM server
func (h *Handler) SCIMHandler(scimHandler http.Handler, routePrefix string) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		rt := h.IntegrationsRuntime
		webhook, err := rt.ResolveWebhookByEndpoint(ctx, endpointID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "unknown SCIM endpoint")
		}

		if subtle.ConstantTimeCompare([]byte(bearerToken), []byte(webhook.SecretToken)) != 1 {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid bearer token")
		}

		installation, err := rt.ResolveIntegration(ctx, integrationsruntime.IntegrationLookup{IntegrationID: webhook.IntegrationID})
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

		ctx = scim.WithRequest(ctx, &scim.Request{
			Installation: installation,
			BasePath:     fmt.Sprintf("%s/%s/v2", routePrefix, endpointID),
		})

		// Strip URL prefix so the SCIM library sees /v2/...
		req := c.Request()
		originalPath := req.URL.Path
		trimPrefix := fmt.Sprintf("%s/%s", routePrefix, endpointID)
		newURL := *req.URL
		newURL.Path = strings.TrimPrefix(originalPath, trimPrefix)

		reqWithPath := req.Clone(ctx)
		reqWithPath.URL = &newURL

		scimHandler.ServeHTTP(c.Response(), reqWithPath)

		return nil
	}
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
