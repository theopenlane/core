package handlers

import (
	"context"
	"strings"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	sso "github.com/theopenlane/core/pkg/ssoutils"
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
func (h *Handler) WebfingerHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.SSOStatusRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
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

			return h.NotFound(ctx, ErrNotFound)
		}

		orgID, err = h.getUserDefaultOrgID(allowCtx, user.ID)
		if err != nil {
			logx.FromContext(reqCtx).Debug().Err(err).Msg("webfinger org lookup failed")

			return h.NotFound(ctx, ErrNotFound)
		}

		userID = user.ID

	default:
		return h.BadRequest(ctx, ErrMissingField)
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	out, err := h.fetchSSOStatus(reqCtx, orgID, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			logx.FromContext(reqCtx).Debug().Err(err).Msg("webfinger org setting not found")

			return h.NotFound(ctx, ErrNotFound)
		}

		logx.FromContext(reqCtx).Error().Err(err).Msg("error fetching sso status")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, out)
}

type nonce string

// fetchSSOStatus returns the SSO status for an organization. For the org level lookup (no userID) it
// reports the raw organization enforcement; when a userID is provided it additionally applies that
// user's owner, per-user, and per-domain exemptions, so callers and the webfinger acct lookup do not
// route an exempt user through SSO. Provider, discovery URL, and TFA enforcement are always reported
func (h *Handler) fetchSSOStatus(ctx context.Context, orgID, userID string) (models.SSOStatusReply, error) {
	in, setting, err := sso.LoadEnforcement(ctx, h.DBClient, orgID, userID, "")
	if err != nil {
		return models.SSOStatusReply{}, err
	}

	out := models.SSOStatusReply{
		Reply:          rout.Reply{Success: true},
		Enforced:       setting.IdentityProviderLoginEnforced,
		OrganizationID: orgID,
		OrgTFAEnforced: setting.MultifactorAuthEnforced,
		IsOrgOwner:     in.IsOwner,
	}

	if userID != "" && out.Enforced {
		out.Enforced = sso.Evaluate(in).MustSSO
	}

	if setting.IdentityProvider != enums.SSOProvider("") {
		out.Provider = setting.IdentityProvider
	}

	if setting.OidcDiscoveryEndpoint != "" {
		out.DiscoveryURL = setting.OidcDiscoveryEndpoint
	}

	return out, nil
}
