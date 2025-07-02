package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"
	"github.com/zitadel/oidc/pkg/client/rp"
	"github.com/zitadel/oidc/pkg/oidc"
	"golang.org/x/oauth2"

	echo "github.com/theopenlane/echox"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/pkg/models"
)

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

// WebfingerHandler determines if SSO login is enforced for an organization or user via a webfinger query
// It parses the resource query param, resolves the org, and returns SSO status
// https://datatracker.ietf.org/doc/html/rfc7033
// confirmed that response codes should not always be 201 or similar, but 404, 200, etc., regular status codes should be used
func (h *Handler) WebfingerHandler(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	resource := ctx.QueryParam("resource")
	if resource == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	var orgID string

	var out models.SSOStatusReply

	switch {
	case strings.HasPrefix(resource, "org:"):
		orgID = strings.TrimPrefix(resource, "org:")
	case strings.HasPrefix(resource, "acct:"):
		email := strings.TrimPrefix(resource, "acct:")

		allowCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

		user, err := h.getUserByEmail(allowCtx, email)
		if err != nil {
			log.Debug().Err(err).Msg("webfinger user lookup failed")

			return h.NotFound(ctx, ErrNotFound)
		}

		orgID, err = h.getUserDefaultOrgID(allowCtx, user.ID)
		if err != nil {
			log.Debug().Err(err).Msg("webfinger org lookup failed")

			return h.NotFound(ctx, ErrNotFound)
		}
	default:
		return h.BadRequest(ctx, ErrMissingField)
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	out, err := h.fetchSSOStatus(reqCtx, orgID)
	if err != nil {
		if ent.IsNotFound(err) {
			log.Debug().Err(err).Msg("webfinger org setting not found")

			return h.NotFound(ctx, ErrNotFound)
		}

		return h.InternalServerError(ctx, err)
	}

	return h.Success(ctx, out)
}

// oidcConfig builds an OIDC relying party config for the given org.
// It loads the org's OIDC settings and constructs the OIDC client config for authentication.
func (h *Handler) oidcConfig(ctx context.Context, orgID string) (rp.RelyingParty, error) {
	// Fetch the organization's OIDC settings from the database
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Ensure all required OIDC config fields are present
	if setting.OidcDiscoveryEndpoint == "" || setting.IdentityProviderClientID == nil || setting.IdentityProviderClientSecret == nil {
		return nil, ErrMissingOIDCConfig
	}

	// Remove the well-known suffix to get the issuer URL
	issuer := strings.TrimSuffix(setting.OidcDiscoveryEndpoint, "/.well-known/openid-configuration")

	// Construct the OIDC relying party configuration
	rpCfg, err := rp.NewRelyingPartyOIDC(
		issuer,                                // OIDC issuer URL
		*setting.IdentityProviderClientID,     // Client ID for the org's IdP
		*setting.IdentityProviderClientSecret, // Client secret for the org's IdP
		h.ssoCallbackURL(),                    // Redirect/callback URL for OIDC flow
		[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail}, // OIDC scopes
		// Configure the nonce verifier to pull the nonce from the context using contextx
		rp.WithVerifierOpts(rp.WithNonce(func(ctx context.Context) string {
			if n, ok := contextx.From[nonce](ctx); ok {
				return string(n)
			}

			return ""
		})),
	)

	if err != nil {
		return nil, err
	}

	return rpCfg, nil
}

// SSOLoginHandler redirects the user to the organization's configured IdP for authentication.
// It sets state and nonce cookies, builds the OIDC auth URL, and issues a redirect.
func (h *Handler) SSOLoginHandler(ctx echo.Context) error {
	orgID := ctx.QueryParam("organization_id")
	if orgID == "" {
		// if no org ID in query, try to get it from cookie
		if c, err := sessions.GetCookie(ctx.Request(), "organization_id"); err == nil {
			orgID = c.Value
		}
	}

	// if a return URL is provided, set it as a cookie for redirect after login
	if ret := ctx.QueryParam("return"); ret != "" {
		sessions.SetCookie(ctx.Response().Writer, ret, "return", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})
	}

	// always set the org ID as a cookie for the OIDC flow
	sessions.SetCookie(ctx.Response().Writer, orgID, "organization_id", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})

	// build the OIDC config for the org
	rpCfg, err := h.oidcConfig(ctx.Request().Context(), orgID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// generate state and nonce for OIDC security
	state := ulids.New().String()
	nonce := ulids.New().String()

	// set state and nonce as cookies for later validation
	sessions.SetCookie(ctx.Response().Writer, state, "state", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})
	sessions.SetCookie(ctx.Response().Writer, nonce, "nonce", sessions.CookieConfig{Path: "/", HTTPOnly: true, SameSite: http.SameSiteLaxMode})

	// build the OIDC auth URL with state and nonce
	authURL := rpCfg.OAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))

	// redirect the user to the IdP for authentication
	return ctx.Redirect(http.StatusFound, authURL)
}

// SSOCallbackHandler completes the OIDC login flow after the user returns from the IdP
// It validates state/nonce, exchanges the code for tokens, provisions the user if needed, and issues a session
func (h *Handler) SSOCallbackHandler(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	orgID := ctx.QueryParam("organization_id")
	if orgID == "" {
		// if no org ID in query, try to get it from cookie
		if c, err := sessions.GetCookie(ctx.Request(), "organization_id"); err == nil {
			orgID = c.Value
		}
	}

	if orgID == "" {
		return h.BadRequest(ctx, ErrMissingField)
	}

	// Build the OIDC config for the org
	rpCfg, err := h.oidcConfig(reqCtx, orgID)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// Validate state matches what was set in the cookie
	stateCookie, err := sessions.GetCookie(ctx.Request(), "state")
	if err != nil || ctx.QueryParam("state") != stateCookie.Value {
		return h.BadRequest(ctx, ErrStateMismatch)
	}

	// Validate nonce exists in the cookie
	nonceCookie, err := sessions.GetCookie(ctx.Request(), "nonce")
	if err != nil {
		return h.BadRequest(ctx, ErrNonceMissing)
	}

	// attach nonce to context for OIDC token validation
	nonceCtx := contextx.With(reqCtx, nonce(nonceCookie.Value))
	// exchange the code for OIDC tokens
	tokens, err := rp.CodeExchange(nonceCtx, ctx.QueryParam("code"), rpCfg)
	if err != nil {
		return h.BadRequest(ctx, err)
	}

	// attach the OIDC email to the context for user provisioning
	ctxWithToken := token.NewContextWithOauthTooToken(reqCtx, tokens.IDTokenClaims.GetEmail())

	// provision the user if they don't exist, or update if they do
	entUser, err := h.CheckAndCreateUser(ctxWithToken, tokens.IDTokenClaims.GetName(), tokens.IDTokenClaims.GetEmail(), enums.AuthProviderOIDC, tokens.IDTokenClaims.GetPicture())
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// set the context for the authenticated user
	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	// build the OAuth session request
	oauthReq := models.OauthTokenRequest{
		Email:            tokens.IDTokenClaims.GetEmail(),
		ExternalUserName: tokens.IDTokenClaims.GetName(),
		AuthProvider:     "oidc",
		Image:            tokens.IDTokenClaims.GetPicture(),
	}

	// generate the session and auth data for the user
	authData, err := h.AuthManager.GenerateOauthAuthSession(userCtx, ctx.Response().Writer, entUser, oauthReq)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	out := models.LoginReply{
		Reply:      rout.Reply{Success: true},
		TFAEnabled: entUser.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *authData,
	}

	// if a return URL was set, redirect there and clean up cookies
	if ret, err := sessions.GetCookie(ctx.Request(), "return"); err == nil && ret.Value != "" {
		sessions.RemoveCookie(ctx.Response().Writer, "return", sessions.CookieConfig{Path: "/"})
		sessions.RemoveCookie(ctx.Response().Writer, "organization_id", sessions.CookieConfig{Path: "/"})

		req, _ := httpsling.Request(httpsling.Get(ret.Value), httpsling.QueryParam("email", tokens.IDTokenClaims.GetEmail()))

		return ctx.Redirect(http.StatusFound, req.URL.String())
	}

	// clean up the org ID cookie after successful login
	sessions.RemoveCookie(ctx.Response().Writer, "organization_id", sessions.CookieConfig{Path: "/"})

	return h.Success(ctx, out)
}

// BindWebfingerHandler documents the webfinger SSO status endpoint for OpenAPI.
func (h *Handler) BindWebfingerHandler() *openapi3.Operation {
	op := openapi3.NewOperation()
	op.Description = "Returns SSO enforcement status for an organization"
	op.OperationID = "WebfingerHandler"
	op.Tags = []string{"authentication"}

	h.AddQueryParameter("resource", op)
	h.AddResponse("SSOStatusReply", "success", models.ExampleSSOStatusReply, op, http.StatusOK)
	op.AddResponse(http.StatusNotFound, notFound())
	op.AddResponse(http.StatusBadRequest, badRequest())
	op.AddResponse(http.StatusInternalServerError, internalServerError())

	return op
}

// ssoCallbackURL builds the callback URL for OIDC flows, ensuring a single path segment is appended.
func (h *Handler) ssoCallbackURL() string {
	base := strings.TrimSuffix(h.OauthProvider.RedirectURL, "/")
	return fmt.Sprintf("%s/v1/sso/callback", base)
}

// ssoOrgForUser checks if the user's default org requires SSO login and the user is not an owner.
// Returns the org ID and true if SSO is enforced and the user must use SSO, otherwise false.
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
		Where(orgmembership.UserID(user.ID), orgmembership.OrganizationID(orgID)).Only(allowCtx)
	if mErr != nil || member.Role == enums.RoleOwner {
		return "", false
	}

	return orgID, true
}

// BindSSOLoginHandler binds the SSO login handler to the OpenAPI schema
func (h *Handler) BindSSOLoginHandler() *openapi3.Operation {
	op := openapi3.NewOperation()
	op.Description = "Initiate the SSO login flow"
	op.OperationID = "SSOLoginHandler"
	op.Tags = []string{"authentication"}
	op.Security = &openapi3.SecurityRequirements{}

	h.AddQueryParameter("organization_id", op)
	op.AddResponse(http.StatusFound, openapi3.NewResponse().WithDescription("Redirect to IdP"))
	op.AddResponse(http.StatusBadRequest, badRequest())

	return op
}

// BindSSOCallbackHandler binds the SSO callback handler to the OpenAPI schema
func (h *Handler) BindSSOCallbackHandler() *openapi3.Operation {
	op := openapi3.NewOperation()
	op.Description = "Complete the OIDC login flow"
	op.OperationID = "SSOCallbackHandler"
	op.Tags = []string{"authentication"}
	op.Security = &openapi3.SecurityRequirements{}

	h.AddQueryParameter("code", op)
	h.AddQueryParameter("state", op)
	h.AddQueryParameter("organization_id", op)
	h.AddResponse("LoginReply", "success", models.ExampleLoginSuccessResponse, op, http.StatusOK)
	op.AddResponse(http.StatusFound, openapi3.NewResponse().WithDescription("Redirect to return URL"))
	op.AddResponse(http.StatusBadRequest, badRequest())
	op.AddResponse(http.StatusInternalServerError, internalServerError())

	return op
}
