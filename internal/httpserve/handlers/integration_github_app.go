package handlers

import (
	"context"
	"encoding/json"
	"net/url"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/githubapp"
	integrationstypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// githubAppSlug is the definition slug for the GitHub App integration
	githubAppSlug = githubapp.Slug

	// GitHub App cookie names used during install callbacks
	githubAppStateCookieName  = "githubapp_state"
	githubAppOrgIDCookieName  = "githubapp_org_id"
	githubAppUserIDCookieName = "githubapp_user_id"
)

// StartGitHubAppInstallation initiates the GitHub App installation flow
func (h *Handler) StartGitHubAppInstallation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	_, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleGitHubAppInstallRequest, openapi.ExampleGitHubAppInstallResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	userCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	if _, err := h.resolveGitHubAppDefinition(); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	state, err := h.generateOAuthState(caller.OrganizationID, githubAppSlug)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Msg("failed to generate github app state")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	installURL := h.githubAppInstallURL(state)

	cfg := h.getOauthCookieConfig()
	sessions.SetCookies(ctx.Response().Writer, cfg, map[string]string{
		githubAppStateCookieName:  state,
		githubAppOrgIDCookieName:  caller.OrganizationID,
		githubAppUserIDCookieName: caller.SubjectID,
	})
	sessions.CopyCookiesFromRequest(ctx.Request(), ctx.Response().Writer, cfg, auth.AccessTokenCookie, auth.RefreshTokenCookie)

	return h.Success(ctx, openapi.GitHubAppInstallResponse{
		Reply:      rout.Reply{Success: true},
		InstallURL: installURL,
		State:      state,
	})
}

// GitHubAppInstallCallback validates callback context and binds installation metadata to the integration.
func (h *Handler) GitHubAppInstallCallback(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	in, err := BindAndValidateQueryParamsWithResponse(ctx, openapiCtx.Operation, openapi.ExampleGitHubAppInstallCallbackRequest, openapi.ExampleGitHubAppInstallCallbackResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}
	if err := h.requireIntegrationsRuntime(ctx, openapiCtx); err != nil {
		return err
	}

	def, err := h.resolveGitHubAppDefinition()
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	req := ctx.Request()
	reqCtx := ctx.Request().Context()

	stateCookie, err := sessions.GetCookie(req, githubAppStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != in.State {
		logx.FromContext(reqCtx).Error().Err(err).Msg("github app state cookie mismatch")
		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	orgID, provider, err := parseStatePayload(in.State)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("invalid github app state payload")
		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	if provider != githubAppSlug {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	logger := logx.FromContext(reqCtx).With().
		Str("installation_id", in.InstallationID).
		Str("org_id", orgID).
		Logger()

	orgCookie, orgErr := sessions.GetCookie(req, githubAppOrgIDCookieName)
	if orgErr != nil || orgCookie.Value == "" || orgCookie.Value != orgID {
		logger.Error().Err(orgErr).Msg("github app callback org id cookie invalid")
		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	userCookie, userErr := sessions.GetCookie(req, githubAppUserIDCookieName)
	if userErr != nil || userCookie.Value == "" {
		logger.Error().Err(userErr).Msg("github app callback user id cookie missing")
		return h.BadRequest(ctx, ErrInvalidUserContext, openapiCtx)
	}

	callbackCaller, callerOk := auth.CallerFromContext(reqCtx)
	if !callerOk || callbackCaller == nil {
		logger.Error().Msg("github app callback has no authenticated user")
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	if err := validateOAuthCallbackIdentity(callbackCaller, orgID, userCookie.Value); err != nil {
		logger.Error().Err(err).Msg("github app callback identity mismatch")
		return h.BadRequest(ctx, err, openapiCtx)
	}

	integrationRecord, err := h.lookupGitHubAppIntegrationByProviderInstallationID(reqCtx, orgID, in.InstallationID)
	if err != nil {
		if !ent.IsNotFound(err) {
			logger.Error().Err(err).Msg("failed to query github app integration")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		integrationRecord, _, err = h.IntegrationsRuntime.EnsureInstallation(reqCtx, orgID, "", def)
		if err != nil {
			logger.Error().Err(err).Msg("failed to resolve github app integration")
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	credential, err := githubapp.MintInstallationCredential(reqCtx, h.IntegrationsConfig.GitHubApp, in.InstallationID)
	if err != nil {
		logger.Error().Err(err).Msg("github app auth failed")
		return h.BadRequest(ctx, err, openapiCtx)
	}

	callbackInput, err := json.Marshal(githubapp.InstallationMetadata{
		InstallationID: in.InstallationID,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshal callback input")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	credentialRegistration, err := resolveCredentialRegistration(def, githubapp.GitHubAppCredential)
	if err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	if err := h.finalizeIntegrationConnection(ctx, openapiCtx, integrationRecord, def, credentialRegistration, credential, callbackInput); err != nil {
		return err
	}

	if err := h.IntegrationsRuntime.SyncWebhooks(reqCtx, integrationRecord, ""); err != nil {
		logger.Error().Err(err).Msg("failed to sync github app webhooks")
	}

	if h.WorkflowEngine == nil || h.Gala == nil {
		logger.Info().Msg("github app vulnerability backfill skipped: workflow engine or gala not configured")
	} else if _, err := h.WorkflowEngine.QueueIntegrationOperation(context.WithoutCancel(reqCtx), engine.IntegrationQueueRequest{
		OrgID:          orgID,
		DefinitionID:   githubapp.DefinitionID.ID(),
		InstallationID: integrationRecord.ID,
		Operation:      githubapp.VulnerabilityCollectOperation.Name(),
		RunType:        enums.IntegrationRunTypeEvent,
	}); err != nil {
		logger.Error().Err(err).Msg("failed to queue github vulnerability backfill")
	}

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, githubAppStateCookieName, githubAppOrgIDCookieName, githubAppUserIDCookieName)

	return h.Success(ctx, openapi.GitHubAppInstallCallbackResponse{Reply: rout.Reply{Success: true}, Message: "GitHub App integration connected"})
}

// githubAppInstallURL builds the GitHub App installation URL including the state parameter.
func (h *Handler) githubAppInstallURL(state string) string {
	installURL := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/apps/" + h.IntegrationsConfig.GitHubApp.AppSlug + "/installations/new",
	}
	if state != "" {
		query := installURL.Query()
		query.Set("state", state)
		installURL.RawQuery = query.Encode()
	}

	return installURL.String()
}

// resolveGitHubAppDefinition validates that the GitHub App provider is active and configured,
// returning the resolved definition on success.
func (h *Handler) resolveGitHubAppDefinition() (integrationstypes.Definition, error) {
	def, ok := h.IntegrationsRuntime.Registry().Definition(githubapp.DefinitionID.ID())
	if !ok || !def.Active {
		return integrationstypes.Definition{}, ErrProviderDisabled
	}

	if h.IntegrationsConfig.GitHubApp.AppSlug == "" {
		return integrationstypes.Definition{}, errGitHubAppNotConfigured
	}

	return def, nil
}
