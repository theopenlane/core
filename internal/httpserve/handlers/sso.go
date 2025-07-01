package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/theopenlane/utils/ulids"
	"github.com/zitadel/oidc/pkg/client/rp"
	"github.com/zitadel/oidc/pkg/oidc"
	"golang.org/x/oauth2"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/models"
)

type nonceKey struct{}

// fetchSSOStatus returns the SSO status for a given organization
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

// WebfingerHandler returns if SSO login is enforced for an organization via a webfinger query
func (h *Handler) WebfingerHandler(ctx echo.Context) error {
	resource := ctx.QueryParam("resource")
	if resource == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	var orgID string
	switch {
	case strings.HasPrefix(resource, "org:"):
		orgID = strings.TrimPrefix(resource, "org:")
	case strings.HasPrefix(resource, "acct:"):
		email := strings.TrimPrefix(resource, "acct:")
		allowCtx := privacy.DecisionContext(ctx.Request().Context(), privacy.Allow)
		user, err := h.getUserByEmail(allowCtx, email)
		if err != nil {
			return h.BadRequest(ctx, err)
		}
		orgID, err = h.getUserDefaultOrgID(allowCtx, user.ID)
		if err != nil {
			return h.BadRequest(ctx, err)
		}
	default:
		return h.BadRequest(ctx, ErrMissingField)
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	out, err := h.fetchSSOStatus(ctx.Request().Context(), orgID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	return h.Success(ctx, out)
}

// oidcConfig builds an oauth2 configuration and oidc provider for an organization.
func (h *Handler) oidcConfig(ctx context.Context, orgID string) (rp.RelyingParty, error) {
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if setting.OidcDiscoveryEndpoint == "" || setting.IdentityProviderClientID == nil || setting.IdentityProviderClientSecret == nil {
		return nil, ErrMissingOIDCConfig
	}

	issuer := strings.TrimSuffix(setting.OidcDiscoveryEndpoint, "/.well-known/openid-configuration")
	rpCfg, err := rp.NewRelyingPartyOIDC(
		issuer,
		*setting.IdentityProviderClientID,
		*setting.IdentityProviderClientSecret,
		h.ssoCallbackURL(),
		[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail},
		rp.WithCustomDiscoveryUrl(setting.OidcDiscoveryEndpoint),
	)
	if err != nil {
		return nil, err
	}

	return rpCfg, nil
}

// SSOLoginHandler redirects the user to the configured IdP for authentication.
func (h *Handler) SSOLoginHandler(ctx echo.Context) error {
	orgID := ctx.QueryParam("organization_id")
	if orgID == "" {
		if c, err := ctx.Request().Cookie("organization_id"); err == nil {
			orgID = c.Value
		}
	}

	if ret := ctx.QueryParam("return"); ret != "" {
		http.SetCookie(ctx.Response(), &http.Cookie{Name: "return", Value: ret, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}

	http.SetCookie(ctx.Response(), &http.Cookie{Name: "organization_id", Value: orgID, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})

	rpCfg, err := h.oidcConfig(ctx.Request().Context(), orgID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	state := ulids.New().String()
	nonce := ulids.New().String()

	http.SetCookie(ctx.Response(), &http.Cookie{Name: "state", Value: state, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
	http.SetCookie(ctx.Response(), &http.Cookie{Name: "nonce", Value: nonce, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})

	authURL := rpCfg.OAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))

	return ctx.Redirect(http.StatusFound, authURL)
}

// SSOCallbackHandler completes the OIDC flow and issues a session for the user.
func (h *Handler) SSOCallbackHandler(ctx echo.Context) error {
	orgID := ctx.QueryParam("organization_id")
	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	rpCfg, err := h.oidcConfig(ctx.Request().Context(), orgID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	stateCookie, err := ctx.Request().Cookie("state")
	if err != nil || ctx.QueryParam("state") != stateCookie.Value {
		return h.BadRequest(ctx, ErrStateMismatch)
	}
	nonceCookie, err := ctx.Request().Cookie("nonce")
	if err != nil {
		return h.BadRequest(ctx, ErrNonceMissing)
	}

	nonceCtx := context.WithValue(ctx.Request().Context(), nonceKey{}, nonceCookie.Value)
	tokens, err := rp.CodeExchange(nonceCtx, ctx.QueryParam("code"), rpCfg)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	ctxWithToken := token.NewContextWithOauthTooToken(ctx.Request().Context(), tokens.IDTokenClaims.GetEmail())

	entUser, err := h.CheckAndCreateUser(ctxWithToken, tokens.IDTokenClaims.GetName(), tokens.IDTokenClaims.GetEmail(), enums.AuthProviderOIDC, tokens.IDTokenClaims.GetPicture())
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	oauthReq := models.OauthTokenRequest{
		Email:            tokens.IDTokenClaims.GetEmail(),
		ExternalUserName: tokens.IDTokenClaims.GetName(),
		AuthProvider:     "oidc",
		Image:            tokens.IDTokenClaims.GetPicture(),
	}

	authData, err := h.AuthManager.GenerateOauthAuthSession(ctxWithToken, ctx.Response().Writer, entUser, oauthReq)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	out := models.LoginReply{
		Reply:      rout.Reply{Success: true},
		TFAEnabled: entUser.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *authData,
	}

	if ret, err := ctx.Request().Cookie("return"); err == nil && ret.Value != "" {
		http.SetCookie(ctx.Response(), &http.Cookie{Name: "return", Value: "", MaxAge: -1, Path: "/"})
		http.SetCookie(ctx.Response(), &http.Cookie{Name: "organization_id", Value: "", MaxAge: -1, Path: "/"})
		q := url.Values{}
		q.Set("email", tokens.IDTokenClaims.GetEmail())
		redirectURL := fmt.Sprintf("%s?%s", ret.Value, q.Encode())
		return ctx.Redirect(http.StatusFound, redirectURL)
	}

	http.SetCookie(ctx.Response(), &http.Cookie{Name: "organization_id", Value: "", MaxAge: -1, Path: "/"})

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

// ssoCallbackURL builds the callback URL for OIDC flows, ensuring there
// is only a single path segment appended to the configured RedirectURL.
func (h *Handler) ssoCallbackURL() string {
	base := strings.TrimSuffix(h.OauthProvider.RedirectURL, "/")
	return fmt.Sprintf("%s/v1/sso/callback", base)
}

// ssoOrgForUser checks if the user's default organization requires SSO login.
// It returns the organization ID when SSO enforcement is active and the user
// is not an owner of that organization.
func (h *Handler) ssoOrgForUser(ctx context.Context, email string) (string, bool) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	user, err := h.getUserByEmail(allowCtx, email)
	if err != nil {
		return "", false
	}

	orgID, err := h.getUserDefaultOrgID(allowCtx, user.ID)
	if err != nil {
		return "", false
	}

	status, err := h.fetchSSOStatus(allowCtx, orgID)
	if err != nil || !status.Enforced {
		return "", false
	}

	member, mErr := transaction.FromContext(allowCtx).OrgMembership.Query().
		Where(
			orgmembership.UserID(user.ID),
			orgmembership.OrganizationID(orgID),
		).Only(allowCtx)
	if mErr != nil || member.Role == enums.RoleOwner {
		return "", false
	}

	return orgID, true
}
