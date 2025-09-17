package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/ulids"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
	apimodels "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/utils/rout"
)

const (
	authenticatedUserSSOCookieName  = "user_sso"
	authenticatedUserSSOCookieValue = "1"
)

// SSOLoginHandler redirects the user to the organization's configured IdP for authentication
// It sets state and nonce cookies, builds the OIDC auth URL, and issues a redirect
// see docs/SSO.md for more details on the SSO flow
func (h *Handler) SSOLoginHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, apimodels.ExampleSSOLoginRequest, rout.Reply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	orgID := in.OrganizationID
	if orgID == "" {
		// if no org ID in query, try to get it from cookie
		if c, err := sessions.GetCookie(ctx.Request(), "organization_id"); err == nil {
			orgID = c.Value
		}
	}

	// if a return URL is provided, set it as a cookie for redirect after login
	cfg := *h.SessionConfig.CookieConfig

	if in.ReturnURL != "" {
		sessions.SetCookie(ctx.Response().Writer, in.ReturnURL, "return", cfg)
	}

	if in.IsTest {
		sessions.SetCookie(ctx.Response().Writer, authenticatedUserSSOCookieValue, authenticatedUserSSOCookieName, cfg)
	}

	authURL, err := h.generateSSOAuthURL(ctx, orgID)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	out := apimodels.SSOLoginReply{
		Reply:       rout.Reply{Success: true},
		RedirectURI: authURL,
	}

	return h.Success(ctx, out, openapi)
}

// SSOCallbackHandler completes the OIDC login flow after the user returns from the IdP
// It validates state/nonce, exchanges the code for tokens, provisions the user if needed, and issues a session
func (h *Handler) SSOCallbackHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	reqCtx := ctx.Request().Context()

	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapi.Operation, apimodels.ExampleSSOCallbackRequest, apimodels.LoginReply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if in.OrganizationID == "" {
		return h.BadRequest(ctx, ErrMissingField, openapi)
	}

	// Build the OIDC config for the org
	rpCfg, err := h.oidcConfig(reqCtx, in.OrganizationID)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// Validate state matches what was set in the cookie
	stateCookie, err := sessions.GetCookie(ctx.Request(), "state")
	if err != nil || in.State != stateCookie.Value {
		return h.BadRequest(ctx, ErrStateMismatch, openapi)
	}

	// Validate nonce exists in the cookie
	nonceCookie, err := sessions.GetCookie(ctx.Request(), "nonce")
	if err != nil {
		return h.BadRequest(ctx, ErrNonceMissing, openapi)
	}

	// attach nonce to context for OIDC token validation
	nonceCtx := contextx.With(reqCtx, nonce(nonceCookie.Value))
	// exchange the code for OIDC tokens
	tokens, err := rp.CodeExchange[*oidc.IDTokenClaims](nonceCtx, in.Code, rpCfg)
	if err != nil {
		return h.BadRequest(ctx, err, openapi)
	}

	// attach the OIDC email to the context for user provisioning
	ctxWithToken := token.NewContextWithOauthTooToken(reqCtx, tokens.IDTokenClaims.Email)

	// provision the user if they don't exist, or update if they do
	entUser, err := h.CheckAndCreateUser(ctxWithToken, tokens.IDTokenClaims.Name, tokens.IDTokenClaims.Email, enums.AuthProviderOIDC, tokens.IDTokenClaims.Picture)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	// set the context for the authenticated user
	userCtx := setAuthenticatedContext(ctxWithToken, entUser)

	// build the OAuth session request
	oauthReq := apimodels.OauthTokenRequest{
		Email:            tokens.IDTokenClaims.Email,
		ExternalUserName: tokens.IDTokenClaims.Name,
		AuthProvider:     "oidc",
		Image:            tokens.IDTokenClaims.Picture,
	}

	// generate the session and auth data for the user
	authData, err := h.AuthManager.GenerateOauthAuthSession(userCtx, ctx.Response().Writer, entUser, oauthReq)
	if err != nil {
		return h.InternalServerError(ctx, err, openapi)
	}

	if tokenID, tErr := sessions.GetCookie(ctx.Request(), "token_id"); tErr == nil {
		tokenType, err := sessions.GetCookie(ctx.Request(), "token_type")
		if err != nil {
			return h.BadRequest(ctx, err)
		}

		orgCookie, err := sessions.GetCookie(ctx.Request(), "organization_id")
		if err != nil {
			return h.BadRequest(ctx, err)
		}

		// we set it on top so will always be here
		user, _ := auth.AuthenticatedUserFromContext(userCtx)

		// make sure every value is set correctly
		user.OrganizationIDs = []string{orgCookie.Value}
		user.OrganizationID = orgCookie.Value

		userCtx = auth.WithAuthenticatedUser(userCtx, user)

		aErr := h.authorizeTokenSSO(privacy.DecisionContext(userCtx, privacy.Allow), tokenType.Value, tokenID.Value, orgCookie.Value)
		if aErr != nil {
			log.Error().Err(aErr).Msg("unable to authorize token for SSO")
		}

		sessions.RemoveCookie(ctx.Response().Writer, "token_id", sessions.CookieConfig{Path: "/"})
		sessions.RemoveCookie(ctx.Response().Writer, "token_type", sessions.CookieConfig{Path: "/"})
	}

	ssoTestCookie, err := sessions.GetCookie(ctx.Request(), authenticatedUserSSOCookieName)
	if err == nil && ssoTestCookie != nil && ssoTestCookie.Value == authenticatedUserSSOCookieValue {
		if err := h.setIDPAuthTested(userCtx, in.OrganizationID); err != nil {
			return err
		}
	}

	out := apimodels.LoginReply{
		Reply:      rout.Reply{Success: true},
		TFAEnabled: entUser.Edges.Setting.IsTfaEnabled,
		Message:    "success",
		AuthData:   *authData,
	}

	// if a return URL was set, redirect there and clean up cookies
	if ret, err := sessions.GetCookie(ctx.Request(), "return"); err == nil && ret.Value != "" {
		sessions.RemoveCookie(ctx.Response().Writer, "return", sessions.CookieConfig{Path: "/"})
		sessions.RemoveCookie(ctx.Response().Writer, "organization_id", sessions.CookieConfig{Path: "/"})

		req, _ := httpsling.Request(httpsling.Get(ret.Value), httpsling.QueryParam("email", tokens.IDTokenClaims.Email))

		return h.Redirect(ctx, req.URL.String(), openapi)
	}

	// clean up the org ID cookie after successful login
	sessions.RemoveCookie(ctx.Response().Writer, "organization_id", sessions.CookieConfig{Path: "/"})

	return h.Success(ctx, out, openapi)
}

// rpConfig holds the configuration for the relying party
type rpConfig struct {
	discovery string
	options   []rp.Option
}

// rpConfigOption defines a function type that modifies the relying party configuration
type rpConfigOption func(*rpConfig)

// withDiscovery sets a custom discovery URL for the relying party configuration
func withDiscovery(url string) rpConfigOption {
	return func(cfg *rpConfig) {
		cfg.discovery = url
	}
}

// withRPOptions allows adding multiple rp.Options to the relying party configuration
func withRPOptions(opts ...rp.Option) rpConfigOption {
	return func(cfg *rpConfig) {
		cfg.options = append(cfg.options, opts...)
	}
}

// newRelyingParty creates a new OIDC relying party instance with the given configuration
func newRelyingParty(ctx context.Context, issuer, clientID, clientSecret, cb string, opts ...rpConfigOption) (rp.RelyingParty, error) {
	cfg := rpConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	rpOpts := cfg.options
	if cfg.discovery != "" {
		rpOpts = append(rpOpts, rp.WithCustomDiscoveryUrl(cfg.discovery))
	}

	return rp.NewRelyingPartyOIDC(
		ctx,
		issuer,
		clientID,
		clientSecret,
		cb,
		[]string{oidc.ScopeOpenID, oidc.ScopeProfile, oidc.ScopeEmail},
		rpOpts...,
	)
}

// setIDPAuthTested makes sure to mark the config as tested so it can be enforced
func (h *Handler) setIDPAuthTested(ctx context.Context, orgID string) error {
	setting, err := h.getOrganizationSettingByOrgID(ctx, orgID)
	if err != nil {
		return err
	}

	if setting.IdentityProviderAuthTested {
		return nil
	}

	return transaction.FromContext(ctx).OrganizationSetting.UpdateOne(setting).
		SetIdentityProviderAuthTested(true).
		Exec(ctx)
}

// oidcConfig builds an OIDC relying party config for the given org.
// to construct the OIDC configuration, the function removes the standard /.well-known/openid-configuration
// suffix from the discovery endpoint to obtain the issuer URL. It then calls rp.NewRelyingPartyOIDC
// to create the relying party instance, passing in the issuer URL, client credentials, the callback URL for the OIDC flow,
// and a set of standard OIDC scopes (openid, profile, email).
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

	// construct the oidc relying party configuration with options
	verifierOpt := rp.WithVerifierOpts(rp.WithNonce(func(ctx context.Context) string {
		if n, ok := contextx.From[nonce](ctx); ok {
			return string(n)
		}

		return ""
	}))

	opts := []rpConfigOption{withRPOptions(verifierOpt)}

	return newRelyingParty(
		ctx,
		issuer,
		*setting.IdentityProviderClientID,
		*setting.IdentityProviderClientSecret,
		h.ssoCallbackURL(),
		append(opts, withDiscovery(setting.OidcDiscoveryEndpoint))...,
	)
}

// ssoCallbackURL builds the callback URL for OIDC flows, ensuring a single path segment is appended
func (h *Handler) ssoCallbackURL() string { return h.OauthProvider.RedirectURL }

// ssoOrgForUser checks if the user's default org requires SSO login and the user is not an owner
// Returns the org ID and true if SSO is enforced and the user must use SSO, otherwise false
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

// authorizeTokenSSO updates the SSO authorization timestamp for a token type (API or Personal Access Token)
func (h *Handler) authorizeTokenSSO(ctx context.Context, tokenType, tokenID, orgID string) error {
	switch tokenType {
	case "api":

		apiToken, err := transaction.FromContext(ctx).APIToken.Get(ctx, tokenID)
		if err != nil {
			return err
		}

		auths := apiToken.SSOAuthorizations
		if auths == nil {
			auths = models.SSOAuthorizationMap{}
		}

		auths[orgID] = time.Now()

		return transaction.FromContext(ctx).APIToken.
			UpdateOne(apiToken).
			SetSSOAuthorizations(auths).
			Exec(ctx)

	case "personal":
		pat, err := transaction.FromContext(ctx).PersonalAccessToken.Get(ctx, tokenID)
		if err != nil {
			return err
		}

		auths := pat.SSOAuthorizations
		if auths == nil {
			auths = models.SSOAuthorizationMap{}
		}

		auths[orgID] = time.Now()

		return transaction.FromContext(ctx).PersonalAccessToken.
			UpdateOne(pat).
			SetSSOAuthorizations(auths).
			Exec(ctx)

	}

	return errInvalidTokenType
}

// generateSSOAuthURL creates an OIDC authentication URL with proper state and nonce cookies
// Returns the authentication URL and any error encountered
//
// The state cookie is used to protect against (CSRF) attacks.
// when the authentication flow is initiated, a unique state value is generated and stored in a cookie.
// later, when the user returns from the identity provider (IdP),
// the application checks that the state value in the callback matches the one stored in the cookie
//
// The nonce cookie is used to prevent replay attacks and to bind the authentication request to the issued ID token.
// the nonce value is sent to the IdP as part of the authentication request, and the IdP includes it in the ID token.
// when the application receives the ID token, it verifies that the nonce matches the one stored in the cookie,
// ensuring the token was issued in response to this specific authentication flow
func (h *Handler) generateSSOAuthURL(ctx echo.Context, orgID string) (string, error) {
	rpCfg, err := h.oidcConfig(ctx.Request().Context(), orgID)
	if err != nil {
		return "", err
	}

	cfg := *h.SessionConfig.CookieConfig

	// set the org ID as a cookie for the OIDC flow
	sessions.SetCookie(ctx.Response().Writer, orgID, "organization_id", cfg)

	state := ulids.New().String()
	nonce := ulids.New().String()

	sessions.SetCookie(ctx.Response().Writer, state, "state", cfg)
	sessions.SetCookie(ctx.Response().Writer, nonce, "nonce", cfg)

	return rpCfg.OAuthConfig().AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce)), nil
}
