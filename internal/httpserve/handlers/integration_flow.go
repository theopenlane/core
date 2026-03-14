package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	oauthStateCookieName  = "oauth_state"
	oauthOrgIDCookieName  = "oauth_org_id"
	oauthUserIDCookieName = "oauth_user_id"
)

// StartOAuthFlow initiates the v2 OAuth flow for an integration definition.
func (h *Handler) StartOAuthFlow(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleOAuthV2FlowRequest, openapi.OAuthFlowResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	definitionID := types.DefinitionID(in.DefinitionID)

	def, ok := h.IntegrationsRuntime.Registry().Definition(definitionID)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Spec.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if def.Auth == nil || def.Auth.Start == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	installationID := in.InstallationID
	if installationID == "" {
		name := def.Spec.DisplayName
		if name == "" {
			name = def.Spec.Slug
		}

		rec, err := h.IntegrationsRuntime.Installations().Create(requestCtx, installation.CreateParams{
			OwnerID:    caller.OrganizationID,
			Name:       name,
			Definition: def.Spec,
		})
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", string(definitionID)).Msg("failed to create installation for oauth flow")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationID = rec.ID
	}

	begin, err := h.IntegrationsRuntime.Keymaker().BeginAuth(requestCtx, keymaker.BeginRequest{
		DefinitionID:   definitionID,
		InstallationID: installationID,
	})
	if err != nil {
		logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", string(definitionID)).Msg("failed to begin oauth flow")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	cfg := h.getOauthCookieConfig()
	sessions.SetCookies(ctx.Response().Writer, cfg, map[string]string{
		oauthOrgIDCookieName:  caller.OrganizationID,
		oauthUserIDCookieName: caller.SubjectID,
		oauthStateCookieName:  begin.State,
	})
	sessions.CopyCookiesFromRequest(ctx.Request(), ctx.Response().Writer, cfg, auth.AccessTokenCookie, auth.RefreshTokenCookie)

	return h.Success(ctx, openapi.OAuthFlowResponse{
		Reply:   rout.Reply{Success: true},
		AuthURL: begin.AuthURL,
		State:   begin.State,
	})
}

// oauthCallbackInput is the JSON payload passed to keymaker.CompleteAuth as the opaque callback input.
type oauthCallbackInput struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// HandleOAuthCallback processes the OAuth callback and persists the resulting credential.
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
	if err != nil || stateCookie.Value == "" {
		logx.FromContext(reqCtx).Error().Err(err).Str("state_fingerprint", stateFingerprint(in.State)).Msg("oauth state cookie not found")

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
		logx.FromContext(reqCtx).Error().Str("cookie_org_id", orgCookie.Value).Str("user_org_id", callbackCaller.OrganizationID).Msg("oauth organization cookie mismatch")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	if callbackCaller.SubjectID != userCookie.Value {
		logx.FromContext(reqCtx).Error().Str("cookie_user_id", userCookie.Value).Str("user_id", callbackCaller.SubjectID).Msg("oauth user cookie mismatch")

		return h.BadRequest(ctx, ErrInvalidUserContext, openapiCtx)
	}

	callbackInput, err := json.Marshal(oauthCallbackInput{Code: in.Code, State: in.State})
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to marshal oauth callback input")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	result, err := h.IntegrationsRuntime.Keymaker().CompleteAuth(reqCtx, keymaker.CompleteRequest{
		State: stateCookie.Value,
		Input: callbackInput,
	})
	if err != nil {
		switch {
		case errors.Is(err, keymaker.ErrAuthStateNotFound),
			errors.Is(err, keymaker.ErrAuthStateExpired),
			errors.Is(err, keymaker.ErrAuthStateTokenRequired):
			return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
		default:
			logx.FromContext(reqCtx).Error().Err(err).Msg("failed to complete oauth callback")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, oauthStateCookieName, oauthOrgIDCookieName, oauthUserIDCookieName)

	def, _ := h.IntegrationsRuntime.Registry().Definition(result.DefinitionID)
	redirectURL := buildV2IntegrationRedirectURL(h.IntegrationsRuntime.SuccessRedirectURL(), def.Spec.Slug)
	if redirectURL == "" {
		return h.Success(ctx, rout.Reply{Success: true})
	}

	return ctx.Redirect(http.StatusFound, redirectURL)
}

// RefreshIntegrationTokenHandler handles requests to refresh an installation's OAuth credential.
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, RefreshInstallationCredentialRequest{}, IntegrationTokenResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	rec, err := h.IntegrationsRuntime.Installations().Get(reqCtx, in.InstallationID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("installation not found")

		return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
	}

	current, ok, err := h.IntegrationsRuntime.CredentialStore().LoadCredential(reqCtx, rec)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("failed to load credential")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if !ok {
		return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
	}

	def, defOk := h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(rec.DefinitionID))
	if !defOk || def.Auth == nil || def.Auth.Refresh == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	refreshed, err := def.Auth.Refresh(reqCtx, current)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("credential refresh failed")

		return h.InternalServerError(ctx, wrapTokenError("refresh", def.Spec.Slug, err), openapiCtx)
	}

	if err := h.IntegrationsRuntime.CredentialStore().SaveInstallationCredential(reqCtx, in.InstallationID, refreshed); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("failed to save refreshed credential")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if refreshed.OAuthAccessToken == "" {
		return h.BadRequest(ctx, wrapTokenError("find access", def.Spec.Slug, ErrIntegrationNotFound), openapiCtx)
	}

	resp := IntegrationTokenResponse{
		Reply:       rout.Reply{Success: true},
		Provider:    def.Spec.Slug,
		AccessToken: refreshed.OAuthAccessToken,
	}

	if refreshed.OAuthExpiry != nil && !refreshed.OAuthExpiry.IsZero() {
		expiry := refreshed.OAuthExpiry.UTC()
		resp.ExpiresAt = &expiry
	}

	return h.Success(ctx, resp)
}

// buildV2IntegrationRedirectURL appends provider and status query params to the base redirect URL.
func buildV2IntegrationRedirectURL(baseURL, slug string) string {
	if strings.TrimSpace(baseURL) == "" {
		return ""
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	q := u.Query()
	q.Set("provider", strings.ToLower(slug))
	q.Set("status", "success")
	q.Set("message", fmt.Sprintf("Successfully connected %s integration", strings.ToLower(slug)))
	u.RawQuery = q.Encode()

	return u.String()
}

// generateOAuthState creates a secure state parameter containing org ID and provider.
// Used by v1 flows (GitHub App, etc.) that manage their own state encoding.
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

// buildIntegrationRedirectURL appends provider and status query params to the base redirect URL.
// Used by v1 flows (GitHub App) that operate on ProviderType.
func buildIntegrationRedirectURL(baseURL string, provider types.ProviderType) string {
	if strings.TrimSpace(baseURL) == "" {
		return ""
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	providerName := strings.ToLower(string(provider))

	q := u.Query()
	q.Set("provider", providerName)
	q.Set("status", "success")
	q.Set("message", fmt.Sprintf("Successfully connected %s integration", providerName))
	u.RawQuery = q.Encode()

	return u.String()
}

// parseProviderType converts a raw provider string to a typed ProviderType.
// Used by v1 flows (GitHub App, disconnect) that still operate on named provider types.
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
