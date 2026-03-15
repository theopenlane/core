package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// githubAppDefinitionID is the canonical definition ID for the GitHub App integration
	githubAppDefinitionID = "def_01K0GHAPP000000000000000001"
	// githubAppSlug is the definition slug for the GitHub App integration
	githubAppSlug = "github_app"
)

// GitHub App cookie names used during install callbacks
const (
	githubAppStateCookieName  = "githubapp_state"
	githubAppOrgIDCookieName  = "githubapp_org_id"
	githubAppUserIDCookieName = "githubapp_user_id"
)

// githubAppAuthInput is the JSON payload passed to Auth.Complete for the GitHub App install flow.
type githubAppAuthInput struct {
	// InstallationID is the GitHub App installation identifier received from the callback
	InstallationID string `json:"installationId"`
}

// StartGitHubAppInstallation initiates the GitHub App installation flow
func (h *Handler) StartGitHubAppInstallation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	_, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleGitHubAppInstallRequest, openapi.ExampleGitHubAppInstallResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	userCtx := ctx.Request().Context()
	caller, ok := auth.CallerFromContext(userCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

	if err := h.validateGitHubAppConfig(); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	state, err := h.generateOAuthState(caller.OrganizationID, githubAppSlug)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Msg("error generating github app state")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	installURL, err := h.githubAppInstallURL(state)
	if err != nil {
		return h.InternalServerError(ctx, err, openapiCtx)
	}

	cfg := h.getOauthCookieConfig()
	sessions.SetCookies(ctx.Response().Writer, cfg, map[string]string{
		githubAppStateCookieName:  state,
		githubAppOrgIDCookieName:  caller.OrganizationID,
		githubAppUserIDCookieName: caller.SubjectID,
	})
	sessions.CopyCookiesFromRequest(ctx.Request(), ctx.Response().Writer, cfg, auth.AccessTokenCookie, auth.RefreshTokenCookie)

	out := openapi.GitHubAppInstallResponse{
		Reply:      rout.Reply{Success: true},
		InstallURL: installURL,
		State:      state,
	}

	return h.Success(ctx, out)
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

	if err := h.validateGitHubAppConfig(); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	req := ctx.Request()
	reqCtx := ctx.Request().Context()

	caller, callerOk := auth.CallerFromContext(reqCtx)
	if !callerOk || caller == nil {
		return h.Unauthorized(ctx, auth.ErrNoAuthUser, openapiCtx)
	}

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

	if caller.OrganizationID != orgID {
		logx.FromContext(reqCtx).Error().Str("state_org_id", orgID).Str("caller_org_id", caller.OrganizationID).Msg("github app callback organization mismatch")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	orgCookie, orgErr := sessions.GetCookie(req, githubAppOrgIDCookieName)
	if orgErr != nil || orgCookie.Value == "" || caller.OrganizationID != orgCookie.Value {
		logx.FromContext(reqCtx).Error().Err(orgErr).Str("caller_org_id", caller.OrganizationID).Msg("github app callback org id cookie invalid")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	userCookie, userErr := sessions.GetCookie(req, githubAppUserIDCookieName)
	if userErr != nil || userCookie.Value == "" || caller.SubjectID != userCookie.Value {
		logx.FromContext(reqCtx).Error().Err(userErr).Str("caller_user_id", caller.SubjectID).Msg("github app callback user id cookie invalid")

		return h.BadRequest(ctx, ErrInvalidUserContext, openapiCtx)
	}

	def, ok := h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(githubAppDefinitionID))
	if !ok || def.Auth == nil || def.Auth.Complete == nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	// Find or create the integration record for this org.
	integrationRecord, err := h.DBClient.Integration.Query().
		Where(
			integration.OwnerID(orgID),
			integration.DefinitionSlugEQ(githubAppSlug),
		).
		Only(reqCtx)
	if err != nil {
		if !ent.IsNotFound(err) {
			logx.FromContext(reqCtx).Error().Err(err).Msg("failed to query github app integration")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}

		integrationRecord, err = h.DBClient.Integration.Create().
			SetOwnerID(orgID).
			SetName("GitHub App").
			SetDefinitionID(githubAppDefinitionID).
			SetDefinitionVersion("v1").
			SetDefinitionSlug(githubAppSlug).
			SetFamily("github").
			SetStatus(enums.IntegrationStatusPending).
			Save(reqCtx)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("failed to create github app integration")

			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	// Complete the auth flow to obtain installation credentials.
	authInput, err := json.Marshal(githubAppAuthInput{InstallationID: in.InstallationID})
	if err != nil {
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	authResult, err := def.Auth.Complete(reqCtx, nil, authInput)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("github app auth completion failed")

		switch {
		case errors.Is(err, ErrProviderHealthCheckFailed):
			return h.BadRequest(ctx, err, openapiCtx)
		case integrationHTTPStatus(err) == http.StatusBadRequest:
			return h.BadRequest(ctx, err, openapiCtx)
		default:
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	// Run health check with the minted credential before persisting.
	if _, err := h.IntegrationsRuntime.Executor().ExecuteOperation(reqCtx, integrationRecord, "health.default", authResult.Credential, nil); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("installation_id", in.InstallationID).Msg("github app health check failed")

		return h.BadRequest(ctx, ErrProviderHealthCheckFailed, openapiCtx)
	}

	// Persist the credential.
	if err := h.IntegrationsRuntime.CredentialStore().SaveCredential(reqCtx, integrationRecord, authResult.Credential); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to save github app credential")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.updateGitHubAppIntegrationMetadata(reqCtx, integrationRecord, authResult.Credential.ProviderData); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to update github app integration metadata")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	h.queueGitHubVulnerabilityBackfill(reqCtx, orgID, integrationRecord.ID)

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, githubAppStateCookieName, githubAppOrgIDCookieName, githubAppUserIDCookieName)

	redirectURL := buildIntegrationRedirectURL(h.IntegrationsRuntime.SuccessRedirectURL(), githubAppSlug)
	if redirectURL == "" {
		return h.Success(ctx, openapi.GitHubAppInstallCallbackResponse{Reply: rout.Reply{Success: true}, Message: "GitHub App integration connected"})
	}

	return h.Redirect(ctx, redirectURL, openapiCtx)
}

// githubAppInstallURL builds the GitHub App installation URL including the state parameter.
func (h *Handler) githubAppInstallURL(state string) (string, error) {
	def, ok := h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(githubAppDefinitionID))
	if !ok || def.Auth == nil || def.Auth.Start == nil {
		return "", ErrProviderDisabled
	}

	result, err := def.Auth.Start(context.Background(), nil)
	if err != nil {
		return "", errGitHubAppNotConfigured
	}

	u, err := url.ParseRequestURI(result.URL)
	if err != nil {
		return "", fmt.Errorf("github app install url parse failed: %w", err)
	}

	q := u.Query()
	q.Set("state", state)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// validateGitHubAppConfig ensures the GitHub App provider is active and all required
// operator credentials are present in the provider spec.
func (h *Handler) validateGitHubAppConfig() error {
	def, ok := h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(githubAppDefinitionID))
	if !ok || !def.Spec.Active {
		return ErrProviderDisabled
	}

	if def.Auth == nil || def.Auth.Start == nil {
		return ErrProviderDisabled
	}

	if _, err := def.Auth.Start(context.Background(), nil); err != nil {
		return errGitHubAppNotConfigured
	}

	return nil
}

// updateGitHubAppIntegrationMetadata merges GitHub App installation metadata into provider state.
func (h *Handler) updateGitHubAppIntegrationMetadata(ctx context.Context, integrationRecord *ent.Integration, providerData json.RawMessage) error {
	nextState := integrationRecord.ProviderState
	if _, err := nextState.MergeProviderData(githubAppSlug, providerData); err != nil {
		return ErrInvalidStateFormat
	}

	return h.DBClient.Integration.UpdateOneID(integrationRecord.ID).
		SetProviderState(nextState).
		Exec(ctx)
}

// resolveOpenlaneOrganizationName returns display_name, then name, then ID
func (h *Handler) resolveOpenlaneOrganizationName(ctx context.Context, orgID string) string {
	org, err := h.DBClient.Organization.Query().Where(organization.ID(orgID)).Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("organization_id", orgID).Msg("failed to resolve openlane organization name")

		return orgID
	}

	if org.DisplayName != "" {
		return org.DisplayName
	}

	if org.Name != "" {
		return org.Name
	}

	return orgID
}

// queueGitHubVulnerabilityBackfill schedules an initial vulnerability collection run after app installation
func (h *Handler) queueGitHubVulnerabilityBackfill(ctx context.Context, orgID, integrationID string) {
	if h.WorkflowEngine == nil {
		logx.FromContext(ctx).Info().Str("org_id", orgID).Msg("github app vulnerability backfill skipped: workflow engine not configured")

		return
	}
	if h.Gala == nil {
		logx.FromContext(ctx).Info().Str("org_id", orgID).Msg("github app vulnerability backfill skipped: gala runtime not configured")

		return
	}

	if _, err := h.WorkflowEngine.QueueIntegrationOperation(context.WithoutCancel(ctx), engine.IntegrationQueueRequest{
		OrgID:          orgID,
		DefinitionID:   githubAppDefinitionID,
		InstallationID: integrationID,
		Operation:      types.OperationName("vulnerability.collect"),
		Force:          true,
		RunType:        enums.IntegrationRunTypeEvent,
	}); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("org_id", orgID).Str("integration_id", integrationID).Msg("failed to queue github vulnerability backfill operation")
	}
}
