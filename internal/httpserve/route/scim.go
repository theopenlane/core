package route

import (
	"fmt"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/httpserve/handlers/scim"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// registerSCIMRoutes sets up SCIM routes
func registerSCIMRoutes(router *Router) error {
	server, err := scim.NewSCIMServer()
	if err != nil {
		return err
	}

	scimHandler := scim.WrapSCIMServerHTTPHandler(server)

	// Integration-scoped SCIM route: /scim/:integrationId/*
	// Validates the integration belongs to the authenticated org,
	// loads SCIMProvisionMode, and injects IntegrationContext before dispatching.
	handler := func(c echo.Context) error {
		req := c.Request()
		ctx := req.Context()

		integrationID := c.PathParam("integrationId")
		if integrationID == "" {
			return echo.ErrBadRequest
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return echo.ErrUnauthorized
		}

		orgID, hasOrg := caller.ActiveOrg()
		if !hasOrg {
			return echo.ErrUnauthorized
		}

		client := transaction.FromContext(ctx)

		integ, err := client.Integration.Query().
			Where(integration.ID(integrationID), integration.OwnerID(orgID)).
			Only(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, scim.ErrIntegrationNotFound.Error())
		}

		mode := integ.Config.SCIMProvisionMode
		if mode == "" {
			mode = enums.SCIMProvisionModeUsers
		}

		ic := &scim.IntegrationContext{
			IntegrationID: integrationID,
			Installation:  integ,
			OrgID:         orgID,
			Runtime:       router.Handler.IntegrationsRuntime,
			ProvisionMode: mode,
		}

		updatedCtx := scim.WithIntegrationContext(ctx, ic)

		// Strip /scim/:integrationId prefix so the SCIM library sees /v2/...
		prefix := fmt.Sprintf("/scim/%s", integrationID)
		newURL := *req.URL
		newURL.Path = strings.TrimPrefix(req.URL.Path, prefix)

		reqWithPath := req.Clone(updatedCtx)
		reqWithPath.URL = &newURL

		scimHandler(c.Response(), reqWithPath)

		return nil
	}

	grp := router.Base()

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	grp.Match(methods, "/scim/:integrationId/*", handler, *authenticatedEndpoint...)

	return nil
}
