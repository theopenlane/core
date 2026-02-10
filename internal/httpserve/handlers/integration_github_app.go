package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/integrations/state"
	"github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/providers/github"
	"github.com/theopenlane/core/pkg/logx"
)

// GitHub App cookie names used during install callbacks.
const (
	githubAppStateCookieName  = "github_app_state"
	githubAppOrgIDCookieName  = "github_app_org_id"
	githubAppUserIDCookieName = "github_app_user_id"
)

// IntegrationGitHubAppConfig contains configuration required to install and operate the GitHub App integration.
type IntegrationGitHubAppConfig struct {
	// Enabled toggles the GitHub App integration handlers.
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// AppID is the GitHub App ID used for JWT signing.
	AppID string `json:"appid" koanf:"appid" default:"" sensitive:"true"`
	// AppSlug is the GitHub App slug used for the install URL.
	AppSlug string `json:"appslug" koanf:"appslug" default:""`
	// PrivateKey is the PEM-encoded GitHub App private key.
	PrivateKey string `json:"privatekey" koanf:"privatekey" default:"" sensitive:"true"`
	// WebhookSecret is the shared secret used to validate GitHub webhooks.
	WebhookSecret string `json:"webhooksecret" koanf:"webhooksecret" default:"" sensitive:"true"`
	// SuccessRedirectURL is the URL to redirect to after successful installation.
	SuccessRedirectURL string `json:"successredirecturl" koanf:"successredirecturl" domain:"inherit" domainPrefix:"https://console" domainSuffix:"/organization-settings/integrations"`
}

// StartGitHubAppInstallation initiates the GitHub App installation flow.
func (h *Handler) StartGitHubAppInstallation(ctx echo.Context, openapiCtx *OpenAPIContext) error {
	_, err := BindAndValidateWithAutoRegistry(ctx, h, openapiCtx.Operation, openapi.ExampleGitHubAppInstallRequest, openapi.ExampleGitHubAppInstallResponse, openapiCtx.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapiCtx)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapiCtx)
	}

	if h.IntegrationStore == nil {
		return h.InternalServerError(ctx, errIntegrationStoreNotConfigured, openapiCtx)
	}

	if err := h.validateGitHubAppConfig(); err != nil {
		return h.BadRequest(ctx, err, openapiCtx)
	}

	_, err = h.IntegrationStore.EnsureIntegration(userCtx, user.OrganizationID, github.TypeGitHubApp)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Str("org_id", user.OrganizationID).Msg("failed to ensure github app integration")
		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	state, err := h.generateOAuthState(user.OrganizationID, string(github.TypeGitHubApp))
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
		githubAppStateCookieName:  state,
		githubAppOrgIDCookieName:  user.OrganizationID,
		githubAppUserIDCookieName: user.SubjectID,
	})

	out := openapi.GitHubAppInstallResponse{
		Reply:      rout.Reply{Success: true},
		InstallURL: installURL,
		State:      state,
	}

	return h.Success(ctx, out, openapiCtx)
}

// GitHubAppInstallCallback finalizes GitHub App installation and stores credentials.
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

	stateCookie, err := sessions.GetCookie(req, githubAppStateCookieName)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != in.State {
		logx.FromContext(reqCtx).Error().Err(err).Msg("github app state cookie mismatch")

		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	orgCookie, err := sessions.GetCookie(req, githubAppOrgIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get github_app_org_id cookie")

		return h.BadRequest(ctx, ErrMissingOrganizationContext, openapiCtx)
	}

	userCookie, err := sessions.GetCookie(req, githubAppUserIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get github_app_user_id cookie")

		return h.BadRequest(ctx, ErrMissingUserContext, openapiCtx)
	}

	orgID, provider, err := parseStatePayload(in.State)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("invalid github app state payload")

		return h.BadRequest(ctx, ErrInvalidState, openapiCtx)
	}

	if provider != string(github.TypeGitHubApp) {
		return h.BadRequest(ctx, ErrInvalidProvider, openapiCtx)
	}

	if orgCookie.Value != orgID {
		logx.FromContext(reqCtx).Error().Str("cookieOrgID", orgCookie.Value).Str("stateOrgID", orgID).Msg("github app organization cookie mismatch")

		return h.BadRequest(ctx, ErrInvalidOrganizationContext, openapiCtx)
	}

	if strings.TrimSpace(userCookie.Value) == "" {
		return h.BadRequest(ctx, ErrMissingUserContext, openapiCtx)
	}

	// Mirror integration_flow behavior: use cookie/state context to set the authenticated user.
	auth.SetAuthenticatedUserContext(ctx, &auth.AuthenticatedUser{
		SubjectID:          userCookie.Value,
		OrganizationID:     orgID,
		OrganizationIDs:    []string{orgID},
		AuthenticationType: auth.JWTAuthentication,
	})
	reqCtx = ctx.Request().Context()

	systemCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	attrs := map[string]any{
		"appId":          strings.TrimSpace(h.IntegrationGitHubApp.AppID),
		"installationId": strings.TrimSpace(in.InstallationID),
		"privateKey":     normalizeGitHubAppPrivateKey(h.IntegrationGitHubApp.PrivateKey),
	}

	if _, err := h.IntegrationStore.EnsureIntegration(systemCtx, orgID, github.TypeGitHubApp); err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("failed to ensure github app integration")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.persistCredentialConfiguration(systemCtx, orgID, github.TypeGitHubApp, attrs); err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("failed to persist github app credentials")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.updateGitHubAppIntegrationMetadata(systemCtx, orgID, attrs); err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("failed to update github app integration metadata")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapiCtx)
	}

	if err := h.runIntegrationHealthCheck(systemCtx, orgID, github.TypeGitHubApp); err != nil {
		logx.FromContext(systemCtx).Error().Err(err).Msg("github app health check failed")
		switch {
		case errors.Is(err, errIntegrationOperationsNotConfigured),
			errors.Is(err, errIntegrationRegistryNotConfigured):
			return h.InternalServerError(ctx, err, openapiCtx)
		default:
			return h.BadRequest(ctx, wrapIntegrationError("validate", err), openapiCtx)
		}
	}

	cfg := h.getOauthCookieConfig()
	h.clearGitHubAppCookies(ctx, cfg)

	redirectURL := buildIntegrationRedirectURL(h.IntegrationGitHubApp.SuccessRedirectURL, github.TypeGitHubApp)
	if redirectURL == "" {
		return h.Success(ctx, openapi.GitHubAppInstallCallbackResponse{Reply: rout.Reply{Success: true}, Message: "GitHub App integration connected"}, openapiCtx)
	}

	return h.Redirect(ctx, redirectURL, openapiCtx)
}

// clearGitHubAppCookies removes cookies used during GitHub App installation.
func (h *Handler) clearGitHubAppCookies(ctx echo.Context, cfg sessions.CookieConfig) {
	writer := ctx.Response().Writer
	for _, name := range []string{
		githubAppStateCookieName,
		githubAppOrgIDCookieName,
		githubAppUserIDCookieName,
	} {
		sessions.RemoveCookie(writer, name, cfg)
	}
}

// validateGitHubAppConfig ensures required GitHub App settings are present.
func (h *Handler) validateGitHubAppConfig() error {
	cfg := h.IntegrationGitHubApp
	if !cfg.Enabled {
		return ErrProviderDisabled
	}
	if strings.TrimSpace(cfg.AppSlug) == "" {
		return rout.MissingField("appSlug")
	}
	if strings.TrimSpace(cfg.AppID) == "" {
		return rout.MissingField("appId")
	}
	if strings.TrimSpace(cfg.PrivateKey) == "" {
		return rout.MissingField("privateKey")
	}
	if strings.TrimSpace(cfg.WebhookSecret) == "" {
		return rout.MissingField("webhookSecret")
	}
	return nil
}

// githubAppInstallURL builds the GitHub App installation URL including the state parameter.
func (h *Handler) githubAppInstallURL(state string) (string, error) {
	slug := strings.TrimSpace(h.IntegrationGitHubApp.AppSlug)
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

// normalizeGitHubAppPrivateKey ensures escaped newlines are converted to PEM newlines.
func normalizeGitHubAppPrivateKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}

	if strings.Contains(trimmed, "\\n") && !strings.Contains(trimmed, "\n") {
		return strings.ReplaceAll(trimmed, "\\n", "\n")
	}

	return trimmed
}

// updateGitHubAppIntegrationMetadata stores the GitHub App installation metadata on the integration record.
func (h *Handler) updateGitHubAppIntegrationMetadata(ctx context.Context, orgID string, attrs map[string]any) error {
	if h == nil || h.DBClient == nil {
		return errDBClientNotConfigured
	}

	appID := strings.TrimSpace(fmt.Sprint(attrs["appId"]))
	installationID := strings.TrimSpace(fmt.Sprint(attrs["installationId"]))
	if appID == "" || installationID == "" {
		return ErrInvalidStateFormat
	}

	statePayload := state.IntegrationProviderState{
		GitHub: &state.GitHubState{
			AppID:          appID,
			InstallationID: installationID,
		},
	}

	return h.DBClient.Integration.Update().
		Where(
			integration.OwnerIDEQ(orgID),
			integration.KindEQ(string(github.TypeGitHubApp)),
		).
		SetProviderState(statePayload).
		Exec(ctx)
}

const defaultHealthOperation types.OperationName = "health.default"

// runIntegrationHealthCheck performs a health check operation for the given provider if supported
func (h *Handler) runIntegrationHealthCheck(ctx context.Context, orgID string, provider types.ProviderType) error {
	if !h.providerHasHealthOperation(provider) {
		return nil
	}

	result, err := h.IntegrationOperations.Run(ctx, types.OperationRequest{
		OrgID:    orgID,
		Provider: provider,
		Name:     defaultHealthOperation,
		Force:    true,
	})
	if err != nil {
		return err
	}

	if result.Status != types.OperationStatusOK {
		summary := strings.TrimSpace(result.Summary)
		if summary == "" {
			return ErrProviderHealthCheckFailed
		}
		return fmt.Errorf("%w: %s", ErrProviderHealthCheckFailed, summary)
	}

	return nil
}

// providerHasHealthOperation checks if the provider has a health check operation defined
func (h *Handler) providerHasHealthOperation(provider types.ProviderType) bool {
	for _, descriptor := range h.IntegrationRegistry.OperationDescriptors(provider) {
		if descriptor.Name == defaultHealthOperation {
			return true
		}
	}

	return false
}
