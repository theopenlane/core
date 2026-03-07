package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	integrationconfig "github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// GitHub App cookie names used during install callbacks
const (
	githubAppStateCookieName  = "githubapp_state"
	githubAppOrgIDCookieName  = "githubapp_org_id"
	githubAppUserIDCookieName = "githubapp_user_id"
)

// githubAppInstallationPayload captures GitHub App installation attributes persisted on the integration.
type githubAppInstallationPayload struct {
	// AppID is the GitHub App identifier
	AppID string `json:"appId"`
	// InstallationID is the installed GitHub App installation identifier
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

	state, err := h.generateOAuthState(caller.OrganizationID, string(github.TypeGitHubApp))
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

	return h.Success(ctx, out, openapiCtx)
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

	if provider != string(github.TypeGitHubApp) {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if caller.OrganizationID != orgID {
		logx.FromContext(reqCtx).Error().Str("state_org_id", orgID).Str("caller_org_id", caller.OrganizationID).Msg("github app callback organization mismatch")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	if orgCookie, orgErr := sessions.GetCookie(req, githubAppOrgIDCookieName); orgErr == nil && orgCookie.Value != "" && caller.OrganizationID != orgCookie.Value {
		logx.FromContext(reqCtx).Error().Str("cookie_org_id", orgCookie.Value).Str("caller_org_id", caller.OrganizationID).Msg("github app callback org cookie mismatch")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	if userCookie, userErr := sessions.GetCookie(req, githubAppUserIDCookieName); userErr == nil && userCookie.Value != "" && caller.SubjectID != userCookie.Value {
		logx.FromContext(reqCtx).Error().Str("cookie_user_id", userCookie.Value).Str("caller_user_id", caller.SubjectID).Msg("github app callback user cookie mismatch")

		return h.BadRequest(ctx, ErrInvalidUserContext, openapiCtx)
	}

	ghSpec, _ := h.gitHubAppSpec()
	installationPayload := githubAppInstallationPayload{AppID: ghSpec.GitHubApp.AppID, InstallationID: in.InstallationID}
	if err := h.verifyGitHubAppInstallation(reqCtx, orgID, installationPayload); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("org_id", orgID).Str("installation_id", in.InstallationID).Msg("github app installation verification failed")

		switch {
		case errors.Is(err, ErrProviderHealthCheckFailed):
			return h.BadRequest(ctx, err, openapiCtx)
		case integrationHTTPStatus(err) == http.StatusBadRequest:
			return h.BadRequest(ctx, err, openapiCtx)
		default:
			return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
		}
	}

	integrationRecord, err := h.IntegrationRuntime.Store().EnsureIntegration(reqCtx, orgID, github.TypeGitHubApp)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to ensure github app integration")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.updateGitHubAppIntegrationMetadata(reqCtx, integrationRecord, installationPayload); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to update github app integration metadata")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	h.queueGitHubVulnerabilityBackfill(reqCtx, orgID, integrationRecord.ID)

	cfg := h.getOauthCookieConfig()
	sessions.RemoveCookies(ctx.Response().Writer, cfg, githubAppStateCookieName, githubAppOrgIDCookieName, githubAppUserIDCookieName)

	redirectURL := buildIntegrationRedirectURL(h.IntegrationRuntime.SuccessRedirectURL(), github.TypeGitHubApp)
	if redirectURL == "" {
		return h.Success(ctx, openapi.GitHubAppInstallCallbackResponse{Reply: rout.Reply{Success: true}, Message: "GitHub App integration connected"}, openapiCtx)
	}

	return h.Redirect(ctx, redirectURL, openapiCtx)
}

// gitHubAppSpec looks up the GitHub App provider spec from the registry.
func (h *Handler) gitHubAppSpec() (integrationconfig.ProviderSpec, bool) {
	spec, ok := h.IntegrationRuntime.Registry().Config(github.TypeGitHubApp)

	return spec, ok
}

// validateGitHubAppConfig ensures the GitHub App provider is active and all required
// operator credentials are present in the provider spec.
func (h *Handler) validateGitHubAppConfig() error {
	spec, ok := h.gitHubAppSpec()
	if !ok || spec.Active == nil || !*spec.Active || spec.GitHubApp == nil {
		return ErrProviderDisabled
	}

	s := spec.GitHubApp
	if s.AppSlug == "" || s.AppID == "" || s.PrivateKey == "" || s.WebhookSecret == "" {
		return errGitHubAppNotConfigured
	}

	return nil
}

// githubAppInstallURL builds the GitHub App installation URL including the state parameter
func (h *Handler) githubAppInstallURL(state string) (string, error) {
	spec, ok := h.gitHubAppSpec()
	if !ok || spec.GitHubApp == nil || spec.GitHubApp.AppSlug == "" {
		return "", rout.MissingField("appSlug")
	}

	slug := spec.GitHubApp.AppSlug

	u := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   fmt.Sprintf("/apps/%s/installations/new", slug),
	}
	q := u.Query()
	q.Set("state", state)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// updateGitHubAppIntegrationMetadata merges GitHub App installation metadata into provider state.
func (h *Handler) updateGitHubAppIntegrationMetadata(ctx context.Context, integrationRecord *ent.Integration, payload githubAppInstallationPayload) error {
	statePatch, err := jsonx.ToMap(payload)
	if err != nil {
		return ErrInvalidStateFormat
	}

	nextState := integrationRecord.ProviderState
	if _, err := nextState.MergeProviderData(string(github.TypeGitHubApp), statePatch); err != nil {
		return ErrInvalidStateFormat
	}

	return h.DBClient.Integration.UpdateOneID(integrationRecord.ID).
		SetProviderState(nextState).
		Exec(ctx)
}

// verifyGitHubAppInstallation mints and validates a GitHub App installation token before persisting installation metadata.
func (h *Handler) verifyGitHubAppInstallation(ctx context.Context, orgID string, installation githubAppInstallationPayload) error {
	credentialPayload, err := types.NewCredentialBuilder(github.TypeGitHubApp).With(
		types.WithCredentialKind(types.CredentialKindMetadata),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: map[string]any{
				"appId":          installation.AppID,
				"installationId": installation.InstallationID,
			},
		}),
	).Build()
	if err != nil {
		return err
	}

	minted, err := h.IntegrationRuntime.Registry().MintPayload(ctx, types.CredentialSubject{
		Provider:   github.TypeGitHubApp,
		OrgID:      orgID,
		Credential: credentialPayload,
	})
	if err != nil {
		return err
	}

	result, err := h.IntegrationRuntime.Operations().RunWithPayload(ctx, types.OperationRequest{
		OrgID:    orgID,
		Provider: github.TypeGitHubApp,
		Name:     types.OperationHealthDefault,
		Force:    true,
	}, minted)
	if err != nil {
		return err
	}
	if result.Status != types.OperationStatusOK {
		return ErrProviderHealthCheckFailed
	}

	return nil
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

	operationName := types.OperationVulnerabilitiesCollect
	if _, err := h.WorkflowEngine.QueueIntegrationOperation(context.WithoutCancel(ctx), engine.IntegrationQueueRequest{
		OrgID:         orgID,
		Provider:      github.TypeGitHubApp,
		IntegrationID: integrationID,
		Operation:     operationName,
		Force:         true,
		RunType:       enums.IntegrationRunTypeEvent,
	}); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("org_id", orgID).Str("integration_id", integrationID).Msg("failed to queue github vulnerability backfill operation")
	}
}
