package handlers

import (
	"context"
	"strings"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
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

	// for an account lookup, reflect the resolved user's owner, per-user, and per-domain exemptions so
	// the client does not route an exempt user through the SSO flow. The org level lookup is unaffected
	if userID != "" && out.Enforced {
		if mustSSO, ssoErr := h.userMustSSO(reqCtx, orgID, userID, ""); ssoErr == nil {
			out.Enforced = mustSSO
		}
	}

	return h.Success(ctx, out, openapi)
}

type nonce string

// fetchSSOStatus returns the organization level SSO enforcement status for a given organization. It
// reports the organization's raw enforcement and TFA settings, the provider, and the discovery URL.
// Per-user exemptions (owner, per-user, per-domain) are intentionally not applied here; that decision
// is resolved separately by userMustSSO at the login, switch, and middleware redirect points
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

// userMustSSO resolves whether the subject must be routed through the SSO login flow for the
// organization, applying owner, per-user, and per-domain exemptions. SSO enforcement status itself is
// reported separately by fetchSSOStatus, which intentionally returns the organization level value
func (h *Handler) userMustSSO(ctx context.Context, orgID, userID, email string) (bool, error) {
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return false, err
	}

	in := sso.EnforcementInput{
		SSOEnforced:   setting.IdentityProviderLoginEnforced,
		ExemptDomains: setting.SSOExemptDomains,
		Email:         email,
	}

	if userID != "" {
		allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

		member, mErr := h.DBClient.OrgMembership.
			Query().Where(
			orgmembership.OrganizationID(orgID),
			orgmembership.UserID(userID),
		).WithUser().Only(allowCtx)
		if mErr == nil {
			in.IsMember = true
			in.IsOwner = member.Role == enums.RoleOwner
			in.MemberExempt = member.SSOExempt

			if in.Email == "" && member.Edges.User != nil {
				in.Email = member.Edges.User.Email
			}
		}
	}

	return sso.Evaluate(in).MustSSO, nil
}
