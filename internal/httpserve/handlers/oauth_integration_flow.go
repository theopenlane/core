package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
	openapi "github.com/theopenlane/core/pkg/openapi"
)

var (
	oauthStateCookieName  = "oauth_state"
	oauthOrgIDCookieName  = "oauth_org_id"
	oauthUserIDCookieName = "oauth_user_id"
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
	} {
		sessions.RemoveCookie(writer, name, cfg)
	}
}

// StartOAuthFlow initiates the OAuth flow for a third-party integration.
func (h *Handler) StartOAuthFlow(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleOAuthFlowRequest, openapi.ExampleOAuthFlowResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
	}

	if h.IntegrationRegistry == nil {
		return h.InternalServerError(ctx, errIntegrationRegistryNotConfigured, openapiCtx)
	}
	if h.KeymakerService == nil {
		return h.InternalServerError(ctx, errKeymakerNotConfigured, openapiCtx)
	}
	if h.IntegrationStore == nil {
		return h.InternalServerError(ctx, errIntegrationStoreNotConfigured, openapiCtx)
	}

	providerType, err := parseProviderType(in.Provider)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	spec, ok := h.IntegrationRegistry.Config(providerType)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}
	if spec.AuthType != types.AuthKindOAuth2 && spec.AuthType != types.AuthKindOIDC {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	integration, err := h.IntegrationStore.EnsureIntegration(userCtx, user.OrganizationID, providerType)
	if err != nil {
		log.Error().Err(err).Str("org_id", user.OrganizationID).Str("provider", string(providerType)).Msg("failed to ensure integration record")
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	state, err := h.generateOAuthState(user.OrganizationID, string(providerType))
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	scopes := mergeScopes(spec, in.Scopes)

	begin, err := h.KeymakerService.BeginAuthorization(userCtx, keymaker.BeginRequest{
		OrgID:         user.OrganizationID,
		IntegrationID: integration.ID,
		Provider:      providerType,
		Scopes:        scopes,
		State:         state,
	})
	if err != nil {
		log.Error().Err(err).Str("provider", string(providerType)).Msg("failed to begin OAuth flow")
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	cfg := h.getOauthCookieConfig()
	h.setOAuthCookies(ctx, cfg, map[string]string{
		oauthOrgIDCookieName:  user.OrganizationID,
		oauthUserIDCookieName: user.SubjectID,
		oauthStateCookieName:  begin.State,
	})

	out := openapi.OAuthFlowResponse{
		Reply:   rout.Reply{Success: true},
		AuthURL: begin.AuthURL,
		State:   begin.State,
	}

	return h.Success(ctx, out)
}

// HandleOAuthCallback processes the OAuth callback and stores integration tokens.
func (h *Handler) HandleOAuthCallback(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.OAuthCallbackRequest{}, openapi.ExampleOAuthCallbackResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	if h.KeymakerService == nil {
		return h.InternalServerError(ctx, errKeymakerNotConfigured, openapiCtx)
	}

	stateCookie, err := sessions.GetCookie(ctx.Request(), oauthStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != in.State {
		log.Error().Err(err).Msg("oauth state cookie mismatch")
		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	orgCookie, err := sessions.GetCookie(ctx.Request(), oauthOrgIDCookieName)
	if err != nil {
		log.Error().Err(err).Msg("failed to get oauth_org_id cookie")
		return h.BadRequest(ctx, ErrMissingOrganizationContext, openapiCtx)
	}

	if _, err := sessions.GetCookie(ctx.Request(), oauthUserIDCookieName); err != nil {
		log.Error().Err(err).Msg("failed to get oauth_user_id cookie")
		return h.BadRequest(ctx, ErrMissingUserContext, openapiCtx)
	}

	result, err := h.KeymakerService.CompleteAuthorization(ctx.Request().Context(), keymaker.CompleteRequest{
		State: in.State,
		Code:  in.Code,
	})
	if err != nil {
		switch {
		case errors.Is(err, integrations.ErrAuthorizationStateNotFound),
			errors.Is(err, integrations.ErrAuthorizationStateExpired),
			errors.Is(err, integrations.ErrAuthorizationCodeRequired):
			return h.BadRequest(ctx, err, openapiCtx)
		default:
			log.Error().Err(err).Msg("failed to complete oauth callback")
			return h.InternalServerError(ctx, err, openapiCtx)
		}
	}

	cfg := h.getOauthCookieConfig()
	h.clearOAuthCookies(ctx, cfg)

	redirectURL := buildIntegrationRedirectURL(h.IntegrationOauthProvider.SuccessRedirectURL, result.Provider)
	if redirectURL == "" {
		return h.Success(ctx, rout.Reply{Success: true})
	}

	log.Info().
		Str("provider", string(result.Provider)).
		Str("org_id", orgCookie.Value).
		Msg("integration oauth flow completed")

	return ctx.Redirect(http.StatusFound, redirectURL)
}

// generateOAuthState creates a secure state parameter containing org ID and provider.
func (h *Handler) generateOAuthState(orgID, provider string) (string, error) {
	const stateRandomBytesLength = 16
	randomBytes := make([]byte, stateRandomBytesLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	stateData := buildStatePayload(orgID, provider, randomBytes)
	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
}

func mergeScopes(spec config.ProviderSpec, requested []string) []string {
	values := map[string]struct{}{}
	add := func(scopes []string) {
		for _, scope := range scopes {
			if trimmed := strings.TrimSpace(scope); trimmed != "" {
				lower := strings.ToLower(trimmed)
				if _, ok := values[lower]; !ok {
					values[lower] = struct{}{}
				}
			}
		}
	}
	if spec.OAuth != nil {
		add(spec.OAuth.Scopes)
	}
	add(requested)

	if len(values) == 0 {
		return nil
	}

	out := make([]string, 0, len(values))
	for scope := range values {
		out = append(out, scope)
	}
	return out
}

func buildIntegrationRedirectURL(baseURL string, provider types.ProviderType) string {
	if strings.TrimSpace(baseURL) == "" {
		return ""
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	providerName := strings.ToLower(string(provider))
	q.Set("provider", strings.ToLower(string(provider)))
	q.Set("status", "success")
	q.Set("message", fmt.Sprintf("Successfully connected %s integration", providerName))
	u.RawQuery = q.Encode()
	return u.String()
}

// retrieveIntegrationToken gets stored OAuth token for a provider.
func (h *Handler) retrieveIntegrationToken(ctx context.Context, orgID, provider string) (*openapi.IntegrationToken, error) {
	if h.IntegrationStore == nil {
		return nil, errIntegrationStoreNotConfigured
	}

	providerType, err := parseProviderType(provider)
	if err != nil {
		return nil, err
	}

	payload, err := h.IntegrationStore.LoadCredential(ctx, orgID, providerType)
	if err != nil {
		return nil, err
	}
	return integrationTokenFromPayload(provider, payload)
}

// RefreshIntegrationToken refreshes an expired OAuth token if refresh token is available.
func (h *Handler) RefreshIntegrationToken(ctx context.Context, orgID, provider string) (*openapi.IntegrationToken, error) {
	if h.IntegrationBroker == nil {
		return nil, wrapTokenError("refresh", provider, errIntegrationBrokerNotConfigured)
	}

	providerType, err := parseProviderType(provider)
	if err != nil {
		return nil, err
	}

	payload, err := h.IntegrationBroker.Mint(ctx, orgID, providerType)
	if err != nil {
		return nil, err
	}
	return integrationTokenFromPayload(provider, payload)
}

// RefreshIntegrationTokenHandler is the HTTP handler for refreshing integration tokens.
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleRefreshIntegrationTokenRequest, openapi.IntegrationTokenResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
	}

	tokenData, err := h.RefreshIntegrationToken(userCtx, user.OrganizationID, in.Provider)
	if err != nil {
		switch {
		case errors.Is(err, keystore.ErrCredentialNotFound):
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", in.Provider, ErrIntegrationNotFound)))
		default:
			return h.InternalServerError(ctx, wrapTokenError("refresh", in.Provider, err), openapiCtx)
		}
	}

	out := openapi.IntegrationTokenResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  in.Provider,
		Token:     tokenData,
		ExpiresAt: tokenData.ExpiresAt,
	}

	return h.Success(ctx, out)
}

func integrationTokenFromPayload(provider string, payload types.CredentialPayload) (*openapi.IntegrationToken, error) {
	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return nil, wrapTokenError("find access", provider, keystore.ErrCredentialNotFound)
	}
	token := tokenOpt.MustGet()

	var expiresAt *time.Time
	if !token.Expiry.IsZero() {
		expiry := token.Expiry
		expiresAt = &expiry
	}

	return &openapi.IntegrationToken{
		Provider:         provider,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		ExpiresAt:        expiresAt,
		ProviderUserID:   "",
		ProviderUsername: "",
		ProviderEmail:    "",
	}, nil
}

func parseProviderType(provider string) (types.ProviderType, error) {
	pt := types.ProviderTypeFromString(provider)
	if pt == types.ProviderUnknown {
		return types.ProviderUnknown, ErrInvalidProvider
	}
	return pt, nil
}

// getOauthCookieConfig returns the cookie configuration for OAuth cookies.
func (h *Handler) getOauthCookieConfig() sessions.CookieConfig {
	secure := !h.IsTest && !h.IsDev

	sameSite := http.SameSiteNoneMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}

	return sessions.CookieConfig{
		Path:     "/",
		HTTPOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
}
