package handlers

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/utils/rout"
)

// WebfingerHandler is a simple protocol which allows you to publicly query a well-know
// URI along with a resource identifier (like an email address) to determine basic attributes
// In our case, we're using it to determine if SSO login is enforced for an organization or user
// It parses the resource query param, resolves the user (or org), and returns SSO status
// https://datatracker.ietf.org/doc/html/rfc7033
// per the RFC, response codes should not always be 201 or similar, but 404, 200, etc.,
// regular status codes should be used
func (h *Handler) WebfingerHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleSSOStatusRequest, models.ExampleSSOStatusReply, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	reqCtx := ctx.Request().Context()

	var orgID string

	switch {
	case strings.HasPrefix(in.Resource, "org:"):
		orgID = strings.TrimPrefix(in.Resource, "org:")
	case strings.HasPrefix(in.Resource, "acct:"):
		email := strings.TrimPrefix(in.Resource, "acct:")

		allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

		user, err := h.getUserByEmail(allowCtx, email)
		if err != nil {
			log.Debug().Err(err).Msg("webfinger user lookup failed")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		orgID, err = h.getUserDefaultOrgID(allowCtx, user.ID)
		if err != nil {
			log.Debug().Err(err).Msg("webfinger org lookup failed")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}
	default:
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	out, err := h.fetchSSOStatus(reqCtx, orgID)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Debug().Err(err).Msg("webfinger org setting not found")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		return h.InternalServerError(ctx, err, openapi)
	}

	return h.Success(ctx, out, openapi)
}

type nonce string

// fetchSSOStatus returns the SSO enforcement status for a given organization
// it checks the organization's settings and returns whether SSO is enforced, the provider, and discovery URL
func (h *Handler) fetchSSOStatus(ctx context.Context, orgID string) (models.SSOStatusReply, error) {
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return models.SSOStatusReply{}, err
	}

	out := models.SSOStatusReply{
		Reply:          rout.Reply{Success: true},
		Enforced:       setting.IdentityProviderLoginEnforced,
		OrganizationID: orgID,
	}

	if setting.IdentityProvider != enums.SSOProvider("") {
		out.Provider = setting.IdentityProvider
	}

	if setting.OidcDiscoveryEndpoint != "" {
		out.DiscoveryURL = setting.OidcDiscoveryEndpoint
	}

	return out, nil
}
