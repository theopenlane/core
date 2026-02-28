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

	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/activation"
	"github.com/theopenlane/core/internal/keystore"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	oauthStateCookieName  = "oauth_state"
	oauthOrgIDCookieName  = "oauth_org_id"
	oauthUserIDCookieName = "oauth_user_id"
)

// StartOAuthFlow initiates the OAuth flow for a third-party integration.
func (h *Handler) StartOAuthFlow(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleOAuthFlowRequest, openapi.ExampleOAuthFlowResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	userCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	providerType, err := parseProviderType(in.Provider)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	spec, specOk := h.IntegrationRegistry.Config(providerType)
	if !specOk {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !lo.FromPtr(spec.Active) {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if spec.AuthType != types.AuthKindOAuth2 && spec.AuthType != types.AuthKindOIDC {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	if h.IntegrationActivation == nil {
		return h.InternalServerError(ctx, errActivationNotConfigured, openapiCtx)
	}

	state, err := h.generateOAuthState(caller.OrganizationID, string(providerType))
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Msg("error generating oauth state")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	scopes := mergeScopes(spec, in.Scopes)

	begin, err := h.IntegrationActivation.BeginOAuth(userCtx, activation.BeginOAuthRequest{
		OrgID:    caller.OrganizationID,
		Provider: providerType,
		Scopes:   scopes,
		State:    state,
	})
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Str("provider", string(providerType)).Msg("failed to begin OAuth flow")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	cfg := h.getOauthCookieConfig()
	sessions.SetCookies(ctx.Response().Writer, cfg, map[string]string{
		oauthOrgIDCookieName:  caller.OrganizationID,
		oauthUserIDCookieName: caller.SubjectID,
		oauthStateCookieName:  begin.State,
	})
	sessions.CopyCookiesFromRequest(ctx.Request(), ctx.Response().Writer, cfg, auth.AccessTokenCookie, auth.RefreshTokenCookie)

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

	if h.IntegrationActivation == nil {
		return h.InternalServerError(ctx, errActivationNotConfigured, openapiCtx)
	}

	reqCtx := ctx.Request().Context()

	stateCookie, err := sessions.GetCookie(ctx.Request(), oauthStateCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("payload_state", in.State).Msg("oauth state cookie not found")
		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	if stateCookie.Value == "" || stateCookie.Value != in.State {
		logx.FromContext(reqCtx).Error().Str("payload_state", in.State).Str("cookie_state", stateCookie.Value).Msg("oauth state cookie mismatch")
		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	orgCookie, err := sessions.GetCookie(ctx.Request(), oauthOrgIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_org_id cookie")
		return h.BadRequest(ctx, ErrMissingOrganizationContext, openapiCtx)
	}

	userCookie, err := sessions.GetCookie(ctx.Request(), oauthUserIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_user_id cookie")
		return h.BadRequest(ctx, ErrMissingUserContext, openapiCtx)
	}

	callbackCaller, callbackOk := auth.CallerFromContext(reqCtx)
	if !callbackOk || callbackCaller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	if callbackCaller.OrganizationID != orgCookie.Value {
		logx.FromContext(reqCtx).Error().Str("cookieOrgID", orgCookie.Value).Str("userOrgID", callbackCaller.OrganizationID).Msg("oauth organization cookie mismatch")
		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	if callbackCaller.SubjectID != userCookie.Value {
		logx.FromContext(reqCtx).Error().Str("cookieUserID", userCookie.Value).Str("userID", callbackCaller.SubjectID).Msg("oauth user cookie mismatch")
		return h.BadRequest(ctx, ErrInvalidUserContext, openapiCtx)
	}

	systemCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	result, err := h.IntegrationActivation.CompleteOAuth(systemCtx, activation.CompleteOAuthRequest{
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
			logx.FromContext(reqCtx).Error().Err(err).Msg("failed to complete oauth callback")
			return h.InternalServerError(ctx, err, openapiCtx)
		}
	}

	integration, err := h.IntegrationStore.EnsureIntegration(systemCtx, result.OrgID, result.Provider)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("org_id", result.OrgID).Str("provider", string(result.Provider)).Msg("failed to ensure integration record")

		return h.InternalServerError(ctx, err, openapiCtx)
	}

	if err := h.updateIntegrationProviderMetadata(systemCtx, integration.ID, result.Provider); err != nil {
		logx.FromContext(reqCtx).Warn().Err(err).Str("provider", string(result.Provider)).Msg("failed to update integration provider metadata")
	}

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, oauthStateCookieName, oauthOrgIDCookieName, oauthUserIDCookieName)

	redirectURL := buildIntegrationRedirectURL(h.IntegrationOauthProvider.SuccessRedirectURL, result.Provider)
	if redirectURL == "" {
		return h.Success(ctx, rout.Reply{Success: true})
	}

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

	return lo.Keys(values)
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
	refreshCaller, refreshOk := auth.CallerFromContext(userCtx)
	if !refreshOk || refreshCaller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	tokenData, err := h.RefreshIntegrationToken(userCtx, refreshCaller.OrganizationID, in.Provider)
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
