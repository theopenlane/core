package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"golang.org/x/oauth2"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/logx"
	models "github.com/theopenlane/core/pkg/openapi"
)

var (
	oauthStateCookieName  = "oauth_state"
	oauthOrgIDCookieName  = "oauth_org_id"
	oauthUserIDCookieName = "oauth_user_id"
	oauthPkceCookieName   = "oauth_pkce"
)

func (h *Handler) setOAuthCookies(ctx echo.Context, cfg sessions.CookieConfig, values map[string]string) {
	writer := ctx.Response().Writer
	for name, value := range values {
		if value == "" {
			continue
		}
		sessions.SetCookie(writer, value, name, cfg)
	}

	if accessCookie, err := sessions.GetCookie(ctx.Request(), auth.AccessTokenCookie); err == nil {
		sessions.SetCookie(writer, accessCookie.Value, auth.AccessTokenCookie, cfg)
	}

	if refreshCookie, err := sessions.GetCookie(ctx.Request(), auth.RefreshTokenCookie); err == nil {
		sessions.SetCookie(writer, refreshCookie.Value, auth.RefreshTokenCookie, cfg)
	}
}

func (h *Handler) clearOAuthCookies(ctx echo.Context, cfg sessions.CookieConfig) {
	writer := ctx.Response().Writer
	for _, name := range []string{
		oauthStateCookieName,
		oauthOrgIDCookieName,
		oauthUserIDCookieName,
		oauthPkceCookieName,
	} {
		sessions.RemoveCookie(writer, name, cfg)
	}
}

func (h *Handler) lookupOAuthProvider(provider string) (*keystore.ProviderRuntime, error) {
	rt, ok := h.IntegrationRegistry[provider]
	if !ok || rt == nil {
		return nil, ErrInvalidProvider
	}
	if rt.Spec.AuthType != keystore.AuthTypeOAuth2 && rt.Spec.AuthType != keystore.AuthTypeOIDC {
		return nil, ErrUnsupportedAuthType
	}
	if rt.OAuthConfig == nil {
		return nil, ErrUnsupportedAuthType
	}
	return rt, nil
}

// StartOAuthFlow initiates the OAuth flow for a third-party integration
func (h *Handler) StartOAuthFlow(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleOAuthFlowRequest, models.ExampleOAuthFlowResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	userCtx := ctx.Request().Context()

	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	// Get user's current organization
	orgID := user.OrganizationID

	// Validate provider via declarative registry
	rt, lookupErr := h.lookupOAuthProvider(in.Provider)
	if lookupErr != nil {
		return h.BadRequest(ctx, lookupErr, openapi)
	}

	// Set up cookie config that will work in either prod, test, or development mode
	cfg := h.getOauthCookieConfig()

	// Generate state parameter for security
	state, err := h.generateOAuthState(orgID, in.Provider)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Msg("failed to generate oauth state")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// Build OAuth authorization URL
	authConfig := *rt.OAuthConfig
	authConfig.Scopes = append([]string{}, authConfig.Scopes...)
	if len(in.Scopes) > 0 {
		// Add additional scopes if requested
		authConfig.Scopes = append(authConfig.Scopes, in.Scopes...)
	}

	authOpts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	cookieValues := map[string]string{
		oauthOrgIDCookieName:  orgID,
		oauthUserIDCookieName: user.SubjectID,
		oauthStateCookieName:  state,
	}
	if rt.Spec.OAuth != nil {
		for key, value := range rt.Spec.OAuth.AuthParams {
			authOpts = append(authOpts, oauth2.SetAuthURLParam(key, value))
		}
		if rt.Spec.OAuth.UsePKCE {
			verifier, err := generatePKCEVerifier()
			if err != nil {
				return h.InternalServerError(ctx, err, openapi)
			}
			challenge := generatePKCEChallenge(verifier)
			authOpts = append(authOpts,
				oauth2.SetAuthURLParam("code_challenge", challenge),
				oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			)
			cookieValues[oauthPkceCookieName] = verifier
		}
	}

	h.setOAuthCookies(ctx, cfg, cookieValues)

	authURL := authConfig.AuthCodeURL(state, authOpts...)

	out := models.OAuthFlowResponse{
		Reply:   rout.Reply{Success: true},
		AuthURL: authURL,
		State:   state,
	}

	return h.Success(ctx, out)
}

// HandleOAuthCallback processes the OAuth callback and stores integration tokens
func (h *Handler) HandleOAuthCallback(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.OAuthCallbackRequest{}, models.ExampleOAuthCallbackResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// Validate state matches what was set in the cookie
	stateCookie, err := sessions.GetCookie(ctx.Request(), oauthStateCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_state cookie")
		return h.BadRequest(ctx, ErrInvalidState, openapi)
	}

	if in.State != stateCookie.Value {
		logx.FromContext(reqCtx).Error().Str("payload state", in.State).Str("cookie state", stateCookie.Value).Msg("State cookies do not match")

		return h.BadRequest(ctx, ErrInvalidState, openapi)
	}

	// Get org ID and user ID from cookies
	orgCookie, err := sessions.GetCookie(ctx.Request(), oauthOrgIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_org_id cookie")

		return h.BadRequest(ctx, ErrMissingOrganizationContext, openapi)
	}

	orgID := orgCookie.Value

	_, err = sessions.GetCookie(ctx.Request(), oauthUserIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_user_id cookie")

		return h.BadRequest(ctx, ErrMissingUserContext, openapi)
	}

	// Get the user from database to set authenticated context
	reqCtx := ctx.Request().Context()
	// Validate state and extract provider from state
	_, provider, err := h.validateOAuthState(in.State)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to validate oauth state")

		return h.BadRequest(ctx, ErrInvalidState, openapi)
	}

	// Set the provider from the state (GitHub doesn't send provider in callback)
	in.Provider = provider

	// Get provider runtime from registry
	rt, lookupErr := h.lookupOAuthProvider(provider)
	if lookupErr != nil {
		return h.BadRequest(ctx, lookupErr, openapi)
	}

	var pkceVerifier string
	if rt.Spec.OAuth != nil && rt.Spec.OAuth.UsePKCE {
		pkceCookie, err := sessions.GetCookie(ctx.Request(), oauthPkceCookieName)
		if err != nil {
			log.Error().Err(err).Msg("failed to get oauth_pkce cookie")
			return h.BadRequest(ctx, ErrInvalidState, openapi)
		}
		pkceVerifier = pkceCookie.Value
		if pkceVerifier == "" {
			log.Error().Msg("pkce verifier cookie empty")
			return h.BadRequest(ctx, ErrInvalidState, openapi)
		}
	}

	// Exchange code for token
	tokenOpts := []oauth2.AuthCodeOption{}
	if rt.Spec.OAuth != nil {
		for key, value := range rt.Spec.OAuth.TokenParams {
			tokenOpts = append(tokenOpts, oauth2.SetAuthURLParam(key, value))
		}
	}
	if pkceVerifier != "" {
		tokenOpts = append(tokenOpts, oauth2.SetAuthURLParam("code_verifier", pkceVerifier))
	}

	oauthToken, err := rt.OAuthConfig.Exchange(reqCtx, in.Code, tokenOpts...)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("provider", provider).Msg("failed to exchange OAuth code for token")

		return h.InternalServerError(ctx, ErrExchangeAuthCode, openapi)
	}

	// Validate token and get user info
	userInfo, err := rt.Validator.Validate(reqCtx, oauthToken.AccessToken, rt)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("provider", provider).Msg("failed to validate OAuth token")

		return h.InternalServerError(ctx, ErrValidateToken, openapi)
	}

	// Store integration and tokens (use authenticated context)
	if _, _, err := h.persistIntegrationTokens(reqCtx, orgID, provider, rt, userInfo, oauthToken); err != nil {
		log.Error().Err(err).Str("provider", provider).Str("org_id", orgID).Msg("failed to store integration tokens")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// Clean up the OAuth cookies after successful completion (like SSO handler)
	cfg := h.getOauthCookieConfig()
	h.clearOAuthCookies(ctx, cfg)

	// Redirect to configured success URL with integration details
	helper := NewIntegrationHelper(provider, "")
	redirectURL := helper.RedirectURL(h.IntegrationOauthProvider.SuccessRedirectURL)

	return ctx.Redirect(http.StatusFound, redirectURL)
}

// generateOAuthState creates a secure state parameter containing org ID and provider
func (h *Handler) generateOAuthState(orgID, provider string) (string, error) {
	// Create random bytes for security
	const stateRandomBytesLength = 16

	randomBytes := make([]byte, stateRandomBytesLength)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Format: orgID:provider:randomBytes (base64 encoded)
	helper := NewIntegrationHelper(provider, "")
	stateData := helper.StateData(orgID, randomBytes)

	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
}

func generatePKCEVerifier() (string, error) {
	const verifierBytes = 32
	b := make([]byte, verifierBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generatePKCEChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// validateOAuthState validates state parameter and extracts org ID and provider
func (h *Handler) validateOAuthState(state string) (orgID, provider string, err error) {
	// Decode base64
	stateBytes, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return "", "", err
	}

	// Split by colons
	parts := strings.Split(string(stateBytes), ":")

	const expectedStateParts = 3
	if len(parts) != expectedStateParts {
		return "", "", ErrInvalidStateFormat
	}

	return parts[0], parts[1], nil
}

func (h *Handler) persistIntegrationTokens(ctx context.Context, orgID, provider string, rt *keystore.ProviderRuntime, userInfo *keystore.IntegrationUserInfo, oauthToken *oauth2.Token) (*ent.Integration, []string, error) {
	if h.IntegrationStore == nil {
		return nil, nil, errIntegrationStoreNotConfigured
	}

	ctxWithOrg := contextx.With(ctx, auth.OrgSubscriptionContextKey{})

	scopes := grantScopes(rt, oauthToken)

	payload := keystore.OAuthTokens{
		OrgID:             orgID,
		Provider:          provider,
		StoreRefreshToken: true,
		AccessToken:       oauthToken.AccessToken,
		RefreshToken:      oauthToken.RefreshToken,
		Scopes:            scopes,
	}

	if userInfo != nil {
		payload.Username = userInfo.Username
		payload.UserID = userInfo.ID
		payload.Email = userInfo.Email
	}

	if rt != nil && rt.Spec.Persistence != nil {
		payload.StoreRefreshToken = rt.Spec.Persistence.StoreRefreshToken
	}

	if !oauthToken.Expiry.IsZero() {
		expiry := oauthToken.Expiry
		payload.ExpiresAt = &expiry
	}

	if userInfo != nil {
		payload.Attributes = map[string]string{}
		if userInfo.ID != "" {
			payload.Attributes[keystore.ProviderUserIDField] = userInfo.ID
		}
		if userInfo.Username != "" {
			payload.Attributes[keystore.ProviderUsernameField] = userInfo.Username
		}
		if userInfo.Email != "" {
			payload.Attributes[keystore.ProviderEmailField] = userInfo.Email
		}
	}

	integration, err := h.IntegrationStore.UpsertOAuthTokens(ctxWithOrg, payload)
	if err != nil {
		return nil, nil, wrapIntegrationError("persist", err)
	}

	return integration, scopes, nil
}

func grantScopes(rt *keystore.ProviderRuntime, token *oauth2.Token) []string {
	values := make(map[string]string)
	add := func(list []string) {
		for _, scope := range list {
			trimmed := strings.TrimSpace(scope)
			if trimmed == "" {
				continue
			}
			values[strings.ToLower(trimmed)] = trimmed
		}
	}
	add(extractTokenScopes(token))
	if rt != nil {
		if rt.OAuthConfig != nil {
			add(rt.OAuthConfig.Scopes)
		}
		if rt.Spec.OAuth != nil {
			add(rt.Spec.OAuth.Scopes)
		}
	}
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, scope := range values {
		out = append(out, scope)
	}
	sort.Strings(out)
	return out
}

func extractTokenScopes(token *oauth2.Token) []string {
	if token == nil {
		return nil
	}
	var scopes []string
	for _, key := range []string{"scope", "scopes"} {
		if raw := token.Extra(key); raw != nil {
			scopes = append(scopes, normalizeScopeValue(raw)...)
		}
	}
	return scopes
}

func normalizeScopeValue(raw any) []string {
	switch v := raw.(type) {
	case string:
		return strings.Fields(v)
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, fmt.Sprint(item))
		}
		return out
	default:
		if raw == nil {
			return nil
		}
		return strings.Fields(fmt.Sprint(raw))
	}
}

// retrieveIntegrationToken gets stored OAuth token for a provider
func (h *Handler) retrieveIntegrationToken(ctx context.Context, orgID, provider string) (*models.IntegrationToken, error) {
	if h.IntegrationStore == nil {
		return nil, errIntegrationStoreNotConfigured
	}

	bundle, err := h.IntegrationStore.LoadTokens(ctx, orgID, provider)
	if err != nil {
		return nil, err
	}

	if bundle.AccessToken == "" {
		return nil, wrapTokenError("find access", provider, ErrIntegrationNotFound)
	}

	return &models.IntegrationToken{
		Provider:         provider,
		AccessToken:      bundle.AccessToken,
		RefreshToken:     bundle.RefreshToken,
		ExpiresAt:        bundle.ExpiresAt,
		ProviderUserID:   bundle.ProviderUserID,
		ProviderUsername: bundle.ProviderUsername,
		ProviderEmail:    bundle.ProviderEmail,
	}, nil
}

// RefreshIntegrationToken refreshes an expired OAuth token if refresh token is available
func (h *Handler) RefreshIntegrationToken(ctx context.Context, orgID, provider string) (*models.IntegrationToken, error) {
	if _, err := h.lookupOAuthProvider(provider); err != nil {
		return nil, err
	}

	if h.IntegrationBroker == nil {
		return nil, wrapTokenError("refresh", provider, errIntegrationBrokerNotConfigured)
	}

	if _, err := h.IntegrationBroker.MintOAuthToken(ctx, orgID, provider); err != nil {
		return nil, wrapTokenError("refresh", provider, err)
	}

	return h.retrieveIntegrationToken(ctx, orgID, provider)
}

// RefreshIntegrationTokenHandler is the HTTP handler for refreshing integration tokens
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleRefreshIntegrationTokenRequest, models.IntegrationTokenResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()

	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	orgID := user.OrganizationID

	// Refresh the token
	tokenData, err := h.RefreshIntegrationToken(userCtx, orgID, in.Provider)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", in.Provider, ErrIntegrationNotFound)))
		}

		return h.InternalServerError(ctx, wrapTokenError("refresh", in.Provider, err), openapi)
	}

	out := models.IntegrationTokenResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  in.Provider,
		Token:     tokenData,
		ExpiresAt: tokenData.ExpiresAt,
	}

	return h.Success(ctx, out)
}

// getOauthCookieConfig returns the cookie configuration for OAuth cookies
// that is dependent on if the environment is test/dev or production
func (h *Handler) getOauthCookieConfig() sessions.CookieConfig {
	secure := !h.IsTest && !h.IsDev

	sameSite := http.SameSiteNoneMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}

	cfg := sessions.CookieConfig{
		Path:     "/",
		HTTPOnly: true,
		Secure:   secure, // Must be true with SameSiteNone in production
		SameSite: sameSite,
	}

	return cfg
}
