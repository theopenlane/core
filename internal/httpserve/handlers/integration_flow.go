package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	oauthStateCookieName  = "oauth_state"
	oauthOrgIDCookieName  = "oauth_org_id"
	oauthUserIDCookieName = "oauth_user_id"
)

// StartOAuthFlow initiates the OAuth flow for an integration definition.
func (h *Handler) StartOAuthFlow(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, ExampleOAuthFlowRequest, openapi.OAuthFlowResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	requestCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(requestCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	def, err := h.resolveActiveDefinition(ctx, in.DefinitionID, openapiCtx)
	if err != nil {
		return err
	}

	if _, err := resolveCredentialRegistration(def, in.CredentialRef); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	if def.Auth == nil || (def.Auth.Start == nil && def.Auth.OAuth == nil) {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	if err := validateDefinitionUserInput(def, in.UserInput); err != nil {
		if errors.Is(err, ErrInvalidInput) {
			return h.InvalidInput(ctx, err, openapiCtx)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	logger := logx.FromContext(requestCtx).With().Str("definition_id", def.ID).Logger()

	installationRec, _, err := h.IntegrationsRuntime.EnsureInstallation(requestCtx, caller.OrganizationID, in.InstallationID, def)
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	if len(in.UserInput) > 0 {
		if err := h.persistInstallationUserInput(requestCtx, installationRec, in.UserInput); err != nil {
			logger.Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to persist user input")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	begin, err := h.IntegrationsRuntime.BeginAuth(requestCtx, keymaker.BeginRequest{
		DefinitionID:   def.ID,
		InstallationID: installationRec.ID,
		CredentialRef:  in.CredentialRef,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to begin oauth flow")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
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
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
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

	if err := validateOAuthCallbackIdentity(callbackCaller, orgCookie.Value, userCookie.Value); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("oauth callback identity mismatch")
		return h.BadRequest(ctx, err, openapiCtx)
	}

	callbackInput, err := json.Marshal(oauthCallbackInput{Code: in.Code, State: in.State})
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to marshal oauth callback input")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if _, err = h.IntegrationsRuntime.CompleteAuth(reqCtx, keymaker.CompleteRequest{
		State: stateCookie.Value,
		Input: callbackInput,
	}); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, oauthStateCookieName, oauthOrgIDCookieName, oauthUserIDCookieName)

	return h.Success(ctx, rout.Reply{Success: true})
}

// RefreshIntegrationTokenHandler handles requests to refresh an installation's OAuth credential.
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, RefreshInstallationCredentialRequest{}, IntegrationTokenResponse{}, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	logger := logx.FromContext(reqCtx).With().Str("installation_id", in.InstallationID).Logger()

	rec, err := h.IntegrationsRuntime.ResolveInstallation(reqCtx, caller.OrganizationID, in.InstallationID, "")
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve installation")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	def, defOk := h.IntegrationsRuntime.Registry().Definition(rec.DefinitionID)
	if !defOk || def.Auth == nil || def.Auth.Refresh == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	credentialRegistration, regErr := resolveCredentialRegistration(def, in.CredentialRef)
	if regErr != nil {
		return h.BadRequest(ctx, regErr, openapiCtx)
	}

	current, ok, err := h.IntegrationsRuntime.LoadCredential(reqCtx, rec, credentialRegistration.Ref)
	if err != nil {
		logger.Error().Err(err).Msg("failed to load credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if !ok {
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	refreshed, err := def.Auth.Refresh(reqCtx, current)
	if err != nil {
		logger.Error().Err(err).Msg("credential refresh failed")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if refreshed.OAuthAccessToken == "" {
		logger.Error().Str("definition_slug", def.Slug).Msg("refreshed credential missing access token")
		return h.BadRequest(ctx, ErrIntegrationNotFound, openapiCtx)
	}

	if err := h.IntegrationsRuntime.SaveInstallationCredential(reqCtx, in.InstallationID, credentialRegistration.Ref, refreshed); err != nil {
		logger.Error().Err(err).Msg("failed to save refreshed credential")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	resp := IntegrationTokenResponse{
		Reply:       rout.Reply{Success: true},
		Provider:    def.Slug,
		AccessToken: refreshed.OAuthAccessToken,
	}

	if refreshed.OAuthExpiry != nil && !refreshed.OAuthExpiry.IsZero() {
		expiry := refreshed.OAuthExpiry.UTC()
		resp.ExpiresAt = &expiry
	}

	return h.Success(ctx, resp)
}

// generateOAuthState creates a secure state parameter containing org ID and provider.
// Used by v1 flows (GitHub App, etc.) that manage their own state encoding.
func (h *Handler) generateOAuthState(orgID, provider string) (string, error) {
	randomPart, err := auth.GenerateOAuthState(stateLength)
	if err != nil {
		return "", err
	}

	stateData := orgID + ":" + provider + ":" + randomPart

	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
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
