package handlers

import (
	"context"
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

	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations"
	integrationspec "github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
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

	spec, specOk := h.IntegrationRuntime.Registry().Config(providerType)
	if !specOk {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !lo.FromPtr(spec.Active) {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if !spec.SupportsInteractiveAuthFlow() {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	integrationID, err := h.resolveOAuthIntegrationID(userCtx, caller.OrganizationID, providerType, in.IntegrationID)
	if err != nil {
		switch {
		case errors.Is(err, ErrIntegrationIDRequired),
			errors.Is(err, ErrIntegrationNotFound),
			errors.Is(err, keystore.ErrIntegrationAmbiguous):
			return h.BadRequest(ctx, err, openapiCtx)
		default:
			logx.FromContext(userCtx).Error().Err(err).Str("provider", string(providerType)).Msg("failed to resolve oauth integration scope")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	state, err := h.generateOAuthState(caller.OrganizationID, string(providerType))
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Msg("error generating oauth state")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	scopes := mergeRequestedScopes(spec, in.Scopes)

	begin, err := h.IntegrationRuntime.Keymaker().BeginAuthorization(userCtx, keymaker.BeginRequest{
		OrgID:         caller.OrganizationID,
		IntegrationID: integrationID,
		Provider:      providerType,
		Scopes:        scopes,
		State:         state,
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

	reqCtx := ctx.Request().Context()

	stateCookie, err := sessions.GetCookie(ctx.Request(), oauthStateCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("state_fingerprint", stateFingerprint(in.State)).Msg("oauth state cookie not found")

		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	if stateCookie.Value == "" || stateCookie.Value != in.State {
		logx.FromContext(reqCtx).Error().Str("state_fingerprint", stateFingerprint(in.State)).Bool("cookie_state_present", stateCookie.Value != "").Msg("oauth state cookie mismatch")

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

	result, err := h.IntegrationRuntime.Keymaker().CompleteAuthorization(systemCtx, keymaker.CompleteRequest{
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

	integrationID := result.IntegrationID
	if integrationID == "" {
		integrationRecord, ensureErr := h.IntegrationRuntime.Store().EnsureIntegration(systemCtx, result.OrgID, result.Provider)
		if ensureErr != nil {
			logx.FromContext(reqCtx).Error().Err(ensureErr).Str("org_id", result.OrgID).Str("provider", string(result.Provider)).Msg("failed to ensure integration record")

			return h.InternalServerError(ctx, ensureErr, openapiCtx)
		}
		integrationID = integrationRecord.ID
	}

	if err := h.updateIntegrationProviderMetadata(systemCtx, integrationID, result.Provider); err != nil {
		logx.FromContext(reqCtx).Warn().Err(err).Str("provider", string(result.Provider)).Msg("failed to update integration provider metadata")
	}

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, oauthStateCookieName, oauthOrgIDCookieName, oauthUserIDCookieName)

	redirectURL := buildIntegrationRedirectURL(h.IntegrationRuntime.SuccessRedirectURL(), result.Provider)
	if redirectURL == "" {
		return h.Success(ctx, rout.Reply{Success: true})
	}

	return ctx.Redirect(http.StatusFound, redirectURL)
}

// generateOAuthState creates a secure state parameter containing org ID and provider.
func (h *Handler) generateOAuthState(orgID, provider string) (string, error) {
	randomPart, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return "", err
	}

	randomBytes, err := base64.RawURLEncoding.DecodeString(randomPart)
	if err != nil {
		return "", ErrInvalidStateFormat
	}

	stateData := buildStatePayload(orgID, provider, randomBytes)

	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
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
func (h *Handler) RefreshIntegrationToken(ctx context.Context, orgID, provider string, integrationID string) (*IntegrationTokenResponse, error) {
	providerType, err := parseProviderType(provider)
	if err != nil {
		return nil, err
	}
	if err := h.validateIntegrationProvider(providerType); err != nil {
		return nil, err
	}

	var payload types.CredentialSet
	if integrationID != "" {
		payload, err = h.IntegrationRuntime.Broker().MintForIntegration(ctx, orgID, providerType, integrationID)
	} else {
		payload, err = h.IntegrationRuntime.Broker().Mint(ctx, orgID, providerType)
	}
	if err != nil {
		return nil, err
	}

	return integrationTokenFromPayload(string(providerType), payload)
}

// RefreshIntegrationTokenHandler is the HTTP handler for refreshing integration tokens.
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleRefreshIntegrationTokenRequest, IntegrationTokenResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	userCtx := ctx.Request().Context()
	refreshCaller, refreshOk := auth.CallerFromContext(userCtx)
	if !refreshOk || refreshCaller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	tokenData, err := h.RefreshIntegrationToken(userCtx, refreshCaller.OrganizationID, in.Provider, in.IntegrationID)
	if err != nil {
		switch integrationHTTPStatus(err) {
		case http.StatusBadRequest:
			return h.BadRequest(ctx, err, openapiCtx)
		default:
			return h.InternalServerError(ctx, wrapTokenError("refresh", in.Provider, err), openapiCtx)
		}
	}

	tokenData.Reply = rout.Reply{Success: true}

	return h.Success(ctx, tokenData)
}

func integrationTokenFromPayload(provider string, payload types.CredentialSet) (*IntegrationTokenResponse, error) {
	if payload.OAuthAccessToken == "" {
		return nil, wrapTokenError("find access", provider, keystore.ErrCredentialNotFound)
	}

	var expiresAt *time.Time
	if payload.OAuthExpiry != nil && !payload.OAuthExpiry.IsZero() {
		expiry := payload.OAuthExpiry.UTC()
		expiresAt = &expiry
	}

	return &IntegrationTokenResponse{
		Provider:    provider,
		AccessToken: payload.OAuthAccessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

func parseProviderType(provider string) (types.ProviderType, error) {
	pt := types.ProviderTypeFromString(provider)
	if pt == types.ProviderUnknown {
		return types.ProviderUnknown, ErrInvalidProvider
	}

	return pt, nil
}

// resolveOAuthIntegrationID resolves which integration record should receive OAuth credentials.
func (h *Handler) resolveOAuthIntegrationID(ctx context.Context, orgID string, provider types.ProviderType, requestedIntegrationID string) (string, error) {
	if requestedIntegrationID != "" {
		if h.DBClient == nil {
			return "", errDBClientNotConfigured
		}

		record, err := h.DBClient.Integration.Query().
			Where(
				integration.IDEQ(requestedIntegrationID),
				integration.OwnerIDEQ(orgID),
				integration.KindEQ(string(provider)),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return "", ErrIntegrationNotFound
			}
			return "", err
		}

		return record.ID, nil
	}

	record, err := h.IntegrationRuntime.Store().EnsureIntegration(ctx, orgID, provider)
	if err != nil {
		if errors.Is(err, keystore.ErrIntegrationAmbiguous) {
			return "", ErrIntegrationIDRequired
		}
		return "", err
	}

	return record.ID, nil
}

// mergeRequestedScopes combines the provider's default scopes with any caller-requested scopes,
// deduplicating by lowercase value. Provider defaults are listed first.
func mergeRequestedScopes(provSpec integrationspec.ProviderSpec, requested []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)

	add := func(scopes []string) {
		for _, scope := range scopes {
			trimmed := strings.TrimSpace(scope)
			if trimmed == "" {
				continue
			}
			lower := strings.ToLower(trimmed)
			if _, ok := seen[lower]; ok {
				continue
			}
			seen[lower] = struct{}{}
			result = append(result, trimmed)
		}
	}

	if provSpec.OAuth != nil {
		add(provSpec.OAuth.Scopes)
	}

	add(requested)

	return result
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
