package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/models"
)

// fetchSSOStatus returns the SSO status for a given organization
func (h *Handler) fetchSSOStatus(ctx context.Context, orgID string) (models.SSOStatusReply, error) {
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return models.SSOStatusReply{}, err
	}

	out := models.SSOStatusReply{
		Reply:    rout.Reply{Success: true},
		Enforced: setting.IdentityProviderLoginEnforced,
	}

	if setting.IdentityProvider != "" {
		out.Provider = setting.IdentityProvider
	}
	if setting.OidcDiscoveryEndpoint != "" {
		out.DiscoveryURL = setting.OidcDiscoveryEndpoint
	}

	return out, nil
}

// WebfingerHandler returns if SSO login is enforced for an organization via a webfinger query
func (h *Handler) WebfingerHandler(ctx echo.Context) error {
	resource := ctx.QueryParam("resource")
	if resource == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	orgID := strings.TrimPrefix(resource, "org:")
	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	out, err := h.fetchSSOStatus(ctx.Request().Context(), orgID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, out)
}

// BindWebfingerHandler binds the webfinger handler to the OpenAPI schema
func (h *Handler) BindWebfingerHandler() *openapi3.Operation {
	op := openapi3.NewOperation()
	op.Description = "Returns SSO enforcement status for an organization"
	op.OperationID = "WebfingerHandler"
	op.Tags = []string{"authentication"}

	h.AddQueryParameter("resource", op)
	h.AddResponse("SSOStatusReply", "success", models.ExampleSSOStatusReply, op, http.StatusOK)
	op.AddResponse(http.StatusBadRequest, badRequest())
	op.AddResponse(http.StatusInternalServerError, internalServerError())

	return op
}
