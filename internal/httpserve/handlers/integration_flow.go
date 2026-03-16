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

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
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

	def, ok := h.IntegrationsRuntime.Registry().Definition(in.DefinitionID)
	if !ok {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if !def.Active {
		return h.BadRequest(ctx, ErrProviderDisabled, openapiCtx)
	}

	if def.Auth == nil || (def.Auth.Start == nil && def.Auth.OAuth == nil) {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	userInputProvided := len(in.UserInput) > 0
	userInput := json.RawMessage(in.UserInput)

	var (
		installationID      string
		installationRec     *ent.Integration
		createdInstallation bool
	)

	if in.InstallationID != "" {
		rec, err := h.IntegrationsRuntime.ResolveInstallation(requestCtx, caller.OrganizationID, in.InstallationID, def.ID)
		if err != nil {
			if errors.Is(err, integrationsruntime.ErrInstallationDefinitionMismatch) {
				return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
			}
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("installation not found")

			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		}

		installationID = rec.ID
		installationRec = rec
	} else {
		name := def.DisplayName
		if name == "" {
			name = def.Slug
		}

		create := h.DBClient.Integration.Create().
			SetOwnerID(caller.OrganizationID).
			SetName(name).
			SetDefinitionID(def.ID).
			SetDefinitionSlug(def.Slug).
			SetFamily(def.Family).
			SetStatus(enums.IntegrationStatusPending)
		if userInputProvided {
			create.SetConfig(types.IntegrationConfig{ClientConfig: userInput})
		}

		rec, err := create.Save(requestCtx)
		if err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", in.DefinitionID).Msg("failed to create installation for oauth flow")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationID = rec.ID
		installationRec = rec
		createdInstallation = true
	}

	if userInputProvided && !createdInstallation {
		config := installationRec.Config
		config.ClientConfig = userInput

		if err := h.DBClient.Integration.UpdateOneID(installationRec.ID).SetConfig(config).Exec(requestCtx); err != nil {
			logx.FromContext(requestCtx).Error().Err(err).Str("installation_id", installationRec.ID).Msg("failed to persist integration user input")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		installationRec.Config = config
	}

	begin, err := h.IntegrationsRuntime.BeginAuth(requestCtx, keymaker.BeginRequest{
		DefinitionID:   def.ID,
		InstallationID: installationID,
	})
	if err != nil {
		if createdInstallation {
			_ = h.DBClient.Integration.DeleteOneID(installationID).Exec(requestCtx)
		}

		switch {
		case errors.Is(err, keymaker.ErrInstallationNotFound),
			errors.Is(err, keymaker.ErrInstallationOwnerMismatch):
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		case errors.Is(err, keymaker.ErrInstallationDefinitionMismatch):
			return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
		default:
			logx.FromContext(requestCtx).Error().Err(err).Str("definition_id", def.ID).Msg("failed to begin oauth flow")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
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

	result, err := h.IntegrationsRuntime.CompleteAuth(reqCtx, keymaker.CompleteRequest{
		State: stateCookie.Value,
		Input: callbackInput,
	})
	if err != nil {
		switch {
		case errors.Is(err, keymaker.ErrAuthStateNotFound),
			errors.Is(err, keymaker.ErrAuthStateExpired),
			errors.Is(err, keymaker.ErrAuthStateTokenRequired):
			return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
		case errors.Is(err, keymaker.ErrInstallationNotFound),
			errors.Is(err, keymaker.ErrInstallationOwnerMismatch):
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		case errors.Is(err, keymaker.ErrInstallationDefinitionMismatch):
			return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
		default:
			logx.FromContext(reqCtx).Error().Err(err).Msg("failed to complete oauth callback")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	if err := h.DBClient.Integration.UpdateOneID(result.InstallationID).
		SetStatus(enums.IntegrationStatusConnected).
		Exec(reqCtx); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", result.InstallationID).Msg("failed to update integration status after oauth callback")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	installationRecord, err := h.IntegrationsRuntime.ResolveInstallation(reqCtx, callbackCaller.OrganizationID, result.InstallationID, result.DefinitionID)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", result.InstallationID).Msg("failed to reload installation after oauth callback")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.IntegrationsRuntime.SyncWebhooks(reqCtx, installationRecord); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", result.InstallationID).Msg("failed to sync integration webhooks after oauth callback")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
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

	rec, err := h.IntegrationsRuntime.ResolveInstallation(reqCtx, caller.OrganizationID, in.InstallationID, "")
	if err != nil {
		switch {
		case errors.Is(err, integrationsruntime.ErrInstallationNotFound):
			logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Str("organization_id", caller.OrganizationID).Msg("installation not found for token refresh")
			return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
		case errors.Is(err, integrationsruntime.ErrInstallationIDRequired):
			return h.BadRequest(ctx, ErrIntegrationIDRequired, openapiCtx)
		default:
			logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Str("organization_id", caller.OrganizationID).Msg("failed to resolve installation for token refresh")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	current, ok, err := h.IntegrationsRuntime.LoadCredential(reqCtx, rec)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("failed to load credential")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if !ok {
		return h.NotFound(ctx, wrapIntegrationError("find", ErrIntegrationNotFound), openapiCtx)
	}

	def, defOk := h.IntegrationsRuntime.Registry().Definition(rec.DefinitionID)
	if !defOk || def.Auth == nil || def.Auth.Refresh == nil {
		return h.BadRequest(ctx, ErrUnsupportedAuthType, openapiCtx)
	}

	refreshed, err := def.Auth.Refresh(reqCtx, current)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("credential refresh failed")

		return h.InternalServerError(ctx, wrapTokenError("refresh", def.Slug, err), openapiCtx)
	}

	if err := h.IntegrationsRuntime.SaveInstallationCredential(reqCtx, in.InstallationID, refreshed); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("failed to save refreshed credential")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if refreshed.OAuthAccessToken == "" {
		return h.BadRequest(ctx, wrapTokenError("find access", def.Slug, ErrIntegrationNotFound), openapiCtx)
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

	randomBytes, err := base64.RawURLEncoding.DecodeString(randomPart)
	if err != nil {
		return "", ErrInvalidStateFormat
	}

	stateData := buildStatePayload(orgID, provider, randomBytes)

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
