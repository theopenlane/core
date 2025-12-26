package handlers

import (
	"context"
	"strings"

	"github.com/theopenlane/common/enums"
	models "github.com/theopenlane/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
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

	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	var orgID, userID string

	switch {
	case strings.HasPrefix(in.Resource, "org:"):
		orgID = strings.TrimPrefix(in.Resource, "org:")
	case strings.HasPrefix(in.Resource, "acct:"):
		email := strings.TrimPrefix(in.Resource, "acct:")

		allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

		user, err := h.getUserByEmail(allowCtx, email)
		if err != nil {
			logx.FromContext(reqCtx).Debug().Err(err).Msg("webfinger user lookup failed")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		orgID, err = h.getUserDefaultOrgID(allowCtx, user.ID)
		if err != nil {
			logx.FromContext(reqCtx).Debug().Err(err).Msg("webfinger org lookup failed")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		userID = user.ID

	default:
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	out, err := h.fetchSSOStatus(reqCtx, orgID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			logx.FromContext(reqCtx).Debug().Err(err).Msg("webfinger org setting not found")

			return h.NotFound(ctx, ErrNotFound, openapi)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error fetching sso status")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return h.Success(ctx, out, openapi)
}

type nonce string

// fetchSSOStatus returns the SSO enforcement status for a given organization
// it checks the organization's settings and returns whether SSO is enforced, the provider, and discovery URL
func (h *Handler) fetchSSOStatus(ctx context.Context, orgID, userID string) (models.SSOStatusReply, error) {
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return models.SSOStatusReply{}, err
	}

	out := models.SSOStatusReply{
		Reply:          rout.Reply{Success: true},
		Enforced:       setting.IdentityProviderLoginEnforced,
		OrganizationID: orgID,
		OrgTFAEnforced: setting.MultifactorAuthEnforced,
	}

	if userID != "" {
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

		member, err := h.DBClient.OrgMembership.
			Query().Where(
			orgmembership.OrganizationID(orgID),
			orgmembership.UserID(userID),
		).Only(allowCtx)
		if err != nil {
			return models.SSOStatusReply{}, err
		}

		out.IsOrgOwner = member.Role == enums.RoleOwner
	}

	if setting.IdentityProvider != enums.SSOProvider("") {
		out.Provider = setting.IdentityProvider
	}

	if setting.OidcDiscoveryEndpoint != "" {
		out.DiscoveryURL = setting.OidcDiscoveryEndpoint
	}

	return out, nil
}
