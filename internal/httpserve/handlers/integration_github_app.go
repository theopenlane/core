package handlers

import (
	"context"
	"fmt"
	"net/url"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// GitHub App cookie names used during install callbacks
const (
	githubAppStateCookieName = "githubapp_state"
)

// IntegrationGitHubAppConfig contains configuration required to install and operate the GitHub App integration
type IntegrationGitHubAppConfig struct {
	// Enabled toggles the GitHub App integration handlers
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// AppID is the GitHub App ID used for JWT signing
	AppID string `json:"appid" koanf:"appid" default:"" sensitive:"true"`
	// AppSlug is the GitHub App slug used for the install URL
	AppSlug string `json:"appslug" koanf:"appslug" default:""`
	// PrivateKey is the PEM-encoded GitHub App private key
	PrivateKey string `json:"privatekey" koanf:"privatekey" default:"" sensitive:"true"`
	// WebhookSecret is the shared secret used to validate GitHub webhooks
	WebhookSecret string `json:"webhooksecret" koanf:"webhooksecret" default:"" sensitive:"true"`
	// SuccessRedirectURL is the URL to redirect to after successful installation
	SuccessRedirectURL string `json:"successredirecturl" koanf:"successredirecturl" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/organization-settings/integrations"`
}

// githubAppInstallationPayload captures GitHub App installation attributes persisted on the integration.
type githubAppInstallationPayload struct {
	// AppID is the GitHub App identifier
	AppID string `json:"appId"`
	// InstallationID is the installed GitHub App installation identifier
	InstallationID string `json:"installationId"`
}

// githubAppProviderStatePatch captures provider state fields persisted on the integration record
type githubAppProviderStatePatch struct {
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

	if h.IntegrationStore == nil {
		return h.InternalServerError(ctx, errIntegrationStoreNotConfigured, openapiCtx)
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
	h.setOAuthCookies(ctx, cfg, map[string]string{
		githubAppStateCookieName: state,
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
	if h.IntegrationStore == nil {
		return h.InternalServerError(ctx, errIntegrationStoreNotConfigured, openapiCtx)
	}

	req := ctx.Request()
	reqCtx := ctx.Request().Context()

	user, err := auth.GetAuthenticatedUserFromContext(reqCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
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

	if user.OrganizationID != orgID {
		logx.FromContext(reqCtx).Error().Str("state_org_id", orgID).Str("user_org_id", user.OrganizationID).Msg("github app callback organization mismatch")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	installationPayload := githubAppInstallationPayload{AppID: h.IntegrationGitHubApp.AppID, InstallationID: in.InstallationID}

	integrationRecord, err := h.IntegrationStore.EnsureIntegration(reqCtx, orgID, github.TypeGitHubApp)
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

	redirectURL := buildIntegrationRedirectURL(h.IntegrationGitHubApp.SuccessRedirectURL, github.TypeGitHubApp)
	if redirectURL == "" {
		return h.Success(ctx, openapi.GitHubAppInstallCallbackResponse{Reply: rout.Reply{Success: true}, Message: "GitHub App integration connected"}, openapiCtx)
	}

	return h.Redirect(ctx, redirectURL, openapiCtx)
}

// validateGitHubAppConfig ensures required GitHub App settings are present.
func (h *Handler) validateGitHubAppConfig() error {
	cfg := h.IntegrationGitHubApp
	if !cfg.Enabled {
		return ErrProviderDisabled
	}
	if cfg.AppSlug == "" {
		return rout.MissingField("appSlug")
	}
	if cfg.AppID == "" {
		return rout.MissingField("appId")
	}
	if cfg.PrivateKey == "" {
		return rout.MissingField("privateKey")
	}
	if cfg.WebhookSecret == "" {
		return rout.MissingField("webhookSecret")
	}

	return nil
}

// githubAppInstallURL builds the GitHub App installation URL including the state parameter
func (h *Handler) githubAppInstallURL(state string) (string, error) {
	slug := h.IntegrationGitHubApp.AppSlug
	if slug == "" {
		return "", rout.MissingField("appSlug")
	}

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
	if h.DBClient == nil {
		return errDBClientNotConfigured
	}

	if integrationRecord == nil || integrationRecord.ID == "" || payload.AppID == "" || payload.InstallationID == "" {
		return ErrInvalidStateFormat
	}

	statePatch, err := jsonx.ToMap(githubAppProviderStatePatch{
		AppID:          payload.AppID,
		InstallationID: payload.InstallationID,
	})
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

// resolveOpenlaneOrganizationName returns display_name, then name, then ID
func (h *Handler) resolveOpenlaneOrganizationName(ctx context.Context, orgID string) string {
	if orgID == "" || h.DBClient == nil {
		return orgID
	}

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
	if orgID == "" || integrationID == "" {
		return
	}

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
