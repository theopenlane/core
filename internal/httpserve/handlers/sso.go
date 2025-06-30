package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/theopenlane/utils/ulids"
	"github.com/zitadel/oidc/pkg/client/rp"
	"github.com/zitadel/oidc/pkg/oidc"
	"golang.org/x/oauth2"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
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
		Reply:    rout.Reply{Success: true},
		Enforced: setting.IdentityProviderLoginEnforced,
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
		fmt.Sprintf("%s/v1/sso/callback", h.OauthProvider.RedirectURL),
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
		return h.BadRequest(ctx, ErrMissingField)
	}

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
