package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
	slackOAuth2 "golang.org/x/oauth2/slack"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	ent "github.com/theopenlane/core/internal/ent/generated"
	hushschema "github.com/theopenlane/core/internal/ent/generated/hush"
	integrationschema "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/models"
)

// IntegrationOauthProviderConfig represents the configuration for OAuth providers used for integrations
type IntegrationOauthProviderConfig struct {
	// RedirectURL is the base URL for integration OAuth callbacks
	RedirectURL string `json:"redirectUrl" koanf:"redirectUrl" default:"http://localhost:17608"`
	// Github contains the configuration settings for GitHub integrations
	Github IntegrationProviderConfig `json:"github" koanf:"github"`
	// Slack contains the configuration settings for Slack integrations
	Slack IntegrationProviderConfig `json:"slack" koanf:"slack"`
}

// IntegrationProviderConfig contains OAuth configuration for a specific integration provider
type IntegrationProviderConfig struct {
	// ClientID is the OAuth2 client ID
	ClientID string `json:"clientId" koanf:"clientId"`
	// ClientSecret is the OAuth2 client secret
	ClientSecret string `json:"clientSecret" koanf:"clientSecret"`
	// ClientEndpoint is the base URL for the OAuth endpoints
	ClientEndpoint string `json:"clientEndpoint" koanf:"clientEndpoint" default:"http://localhost:17608"`
	// Scopes are the OAuth2 scopes to request
	Scopes []string `json:"scopes" koanf:"scopes"`
}

// IntegrationProvider represents a supported OAuth provider for integrations
type IntegrationProvider struct {
	Name     string
	Config   *oauth2.Config
	Validate func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error)
}

// IntegrationUserInfo contains user information from OAuth provider
type IntegrationUserInfo struct {
	ID       string
	Username string
	Email    string
}

var (
	// ErrInvalidState is returned when OAuth state validation fails
	ErrInvalidState = fmt.Errorf("invalid OAuth state parameter")
	// ErrMissingCode is returned when OAuth authorization code is missing
	ErrMissingCode = fmt.Errorf("missing OAuth authorization code")
	// ErrExchangeAuthCode is returned when OAuth code exchange fails
	ErrExchangeAuthCode = fmt.Errorf("failed to exchange authorization code")
	// ErrValidateToken is returned when OAuth token validation fails
	ErrValidateToken = fmt.Errorf("failed to validate OAuth token")
	// ErrInvalidStateFormat is returned when OAuth state format is invalid
	ErrInvalidStateFormat = fmt.Errorf("invalid state format")
	// ErrProviderRequired is returned when provider parameter is missing
	ErrProviderRequired = fmt.Errorf("provider parameter is required")
	// ErrIntegrationIDRequired is returned when integration ID is missing
	ErrIntegrationIDRequired = fmt.Errorf("integration ID is required")
	// ErrIntegrationNotFound is returned when integration is not found
	ErrIntegrationNotFound = fmt.Errorf("integration not found")
	// ErrDeleteSecrets is returned when deleting integration secrets fails
	ErrDeleteSecrets = fmt.Errorf("failed to delete integration secrets")
)

// getIntegrationProviders returns available OAuth providers for integrations
func (h *Handler) getIntegrationProviders() map[string]IntegrationProvider {
	providers := map[string]IntegrationProvider{
		"github": {
			Name:   "github",
			Config: h.getGithubIntegrationConfig(),
			Validate: func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
				return h.validateGithubIntegrationToken(ctx, token)
			},
		},
	}

	// Only add Slack if configuration is available
	if slackConfig := h.getSlackIntegrationConfig(); slackConfig != nil {
		providers["slack"] = IntegrationProvider{
			Name:   "slack",
			Config: slackConfig,
			Validate: func(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
				return h.validateSlackIntegrationToken(ctx, token)
			},
		}
	}

	return providers
}

// getGithubIntegrationConfig returns OAuth2 config for GitHub integrations
func (h *Handler) getGithubIntegrationConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.IntegrationOauthProvider.Github.ClientID,
		ClientSecret: h.IntegrationOauthProvider.Github.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/v1/integrations/oauth/callback", h.IntegrationOauthProvider.Github.ClientEndpoint),
		Endpoint:     githubOAuth2.Endpoint,
		Scopes:       []string{"read:user", "user:email", "repo"}, // Extended scopes for API access
	}
}

// getSlackIntegrationConfig returns OAuth2 config for Slack integrations
func (h *Handler) getSlackIntegrationConfig() *oauth2.Config {
	// Check if Slack configuration is available
	if h.IntegrationOauthProvider.Slack.ClientID == "" || h.IntegrationOauthProvider.Slack.ClientSecret == "" {
		return nil
	}

	return &oauth2.Config{
		ClientID:     h.IntegrationOauthProvider.Slack.ClientID,
		ClientSecret: h.IntegrationOauthProvider.Slack.ClientSecret,
		RedirectURL:  fmt.Sprintf("%s/v1/integrations/oauth/callback", h.IntegrationOauthProvider.Slack.ClientEndpoint),
		Endpoint:     slackOAuth2.Endpoint,
		Scopes:       h.IntegrationOauthProvider.Slack.Scopes, // Default Slack scopes for integrations
	}
}

// StartOAuthFlow initiates the OAuth flow for a third-party integration
func (h *Handler) StartOAuthFlow(ctx echo.Context) error {
	var in models.OAuthFlowRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// Get the authenticated user (authMW ensures user is authenticated)
	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		// Debug: log authentication failure details
		log.Error().
			Err(err).
			Str("provider", in.Provider).
			Msg("Failed to get authenticated user from context for OAuth integration")
		return h.Unauthorized(ctx, err)
	}

	// Debug: log successful authentication
	log.Info().
		Str("user_id", user.SubjectID).
		Str("org_id", user.OrganizationID).
		Str("provider", in.Provider).
		Msg("Starting OAuth integration flow for authenticated user")

	// Get user's current organization
	orgID := user.OrganizationID

	// Validate provider
	providers := h.getIntegrationProviders()
	provider, exists := providers[strings.ToLower(in.Provider)]
	if !exists {
		return h.BadRequest(ctx, ErrInvalidProvider)
	}

	// Set up cookie config with SameSiteNoneMode for OAuth flow to work with external redirects
	cfg := sessions.CookieConfig{
		Path:     "/",
		HTTPOnly: true,
		SameSite: http.SameSiteNoneMode, // Required for OAuth redirects from external providers
		Secure:   !h.IsTest,             // Must be true with SameSiteNone in production
	}

	// Set the org ID and user ID as cookies for the OAuth flow
	sessions.SetCookie(ctx.Response().Writer, orgID, "oauth_org_id", cfg)
	sessions.SetCookie(ctx.Response().Writer, user.SubjectID, "oauth_user_id", cfg)

	// Re-set existing auth cookies with SameSiteNone for OAuth compatibility
	if accessCookie, err := sessions.GetCookie(ctx.Request(), auth.AccessTokenCookie); err == nil {
		sessions.SetCookie(ctx.Response().Writer, accessCookie.Value, auth.AccessTokenCookie, cfg)
	}
	if refreshCookie, err := sessions.GetCookie(ctx.Request(), auth.RefreshTokenCookie); err == nil {
		sessions.SetCookie(ctx.Response().Writer, refreshCookie.Value, auth.RefreshTokenCookie, cfg)
	}

	// Generate state parameter for security
	state, err := h.generateOAuthState(orgID, in.Provider)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	// Set the state as a cookie for validation in callback
	sessions.SetCookie(ctx.Response().Writer, state, "oauth_state", cfg)

	// Build OAuth authorization URL
	config := provider.Config
	if len(in.Scopes) > 0 {
		// Add additional scopes if requested
		config = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Endpoint:     config.Endpoint,
			Scopes:       append(config.Scopes, in.Scopes...),
		}
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	out := models.OAuthFlowResponse{
		Reply:   rout.Reply{Success: true},
		AuthURL: authURL,
		State:   state,
	}

	return h.Success(ctx, out)
}

// HandleOAuthCallback processes the OAuth callback and stores integration tokens
func (h *Handler) HandleOAuthCallback(ctx echo.Context) error {
	var in models.OAuthCallbackRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	if err := in.Validate(); err != nil {
		return h.InvalidInput(ctx, err)
	}

	// Validate state matches what was set in the cookie (like SSO handler)
	stateCookie, err := sessions.GetCookie(ctx.Request(), "oauth_state")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get oauth_state cookie")
		return h.BadRequest(ctx, ErrInvalidState)
	}
	if in.State != stateCookie.Value {
		log.Error().
			Str("expected_state", stateCookie.Value).
			Str("received_state", in.State).
			Msg("OAuth state mismatch")
		return h.BadRequest(ctx, ErrInvalidState)
	}

	// Get org ID and user ID from cookies (like SSO handler)
	orgCookie, err := sessions.GetCookie(ctx.Request(), "oauth_org_id")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get oauth_org_id cookie")
		return h.BadRequest(ctx, fmt.Errorf("missing organization context: %w", ErrInvalidState))
	}
	orgID := orgCookie.Value

	userCookie, err := sessions.GetCookie(ctx.Request(), "oauth_user_id")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get oauth_user_id cookie")
		return h.BadRequest(ctx, fmt.Errorf("missing user context: %w", ErrInvalidState))
	}
	_ = userCookie.Value

	// Get the user from database to set authenticated context (like SSO handler)
	reqCtx := ctx.Request().Context()
	systemCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	// Validate state and extract provider from state
	_, provider, err := h.validateOAuthState(in.State)
	if err != nil {
		return h.BadRequest(ctx, ErrInvalidState)
	}

	// Set the provider from the state (GitHub doesn't send provider in callback)
	in.Provider = provider

	// Get provider configuration
	providers := h.getIntegrationProviders()
	providerConfig, exists := providers[provider]
	if !exists {
		return h.BadRequest(ctx, ErrInvalidProvider)
	}

	// Exchange code for token
	oauthToken, err := providerConfig.Config.Exchange(ctx.Request().Context(), in.Code)
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("failed to exchange OAuth code for token")
		return h.InternalServerError(ctx, ErrExchangeAuthCode)
	}

	// Validate token and get user info
	userInfo, err := providerConfig.Validate(ctx.Request().Context(), oauthToken)
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("failed to validate OAuth token")
		return h.InternalServerError(ctx, ErrValidateToken)
	}

	// Store integration and tokens (use authenticated context)
	_, err = h.storeIntegrationTokens(systemCtx, orgID, provider, userInfo, oauthToken)
	if err != nil {
		log.Error().Err(err).Str("provider", provider).Str("org_id", orgID).Msg("failed to store integration tokens")
		return h.InternalServerError(ctx, err)
	}

	// Clean up the OAuth cookies after successful completion (like SSO handler)
	sessions.RemoveCookie(ctx.Response().Writer, "oauth_state", sessions.CookieConfig{Path: "/"})
	sessions.RemoveCookie(ctx.Response().Writer, "oauth_org_id", sessions.CookieConfig{Path: "/"})
	sessions.RemoveCookie(ctx.Response().Writer, "oauth_user_id", sessions.CookieConfig{Path: "/"})

	// Redirect back to the HTML interface with success message
	redirectURL := fmt.Sprintf("/pkg/testutils/integrations/index.html?provider=%s&status=success&message=%s",
		provider,
		fmt.Sprintf("Successfully connected %s integration", cases.Title(language.English).String(provider)))

	return ctx.Redirect(http.StatusFound, redirectURL)
}

// generateOAuthState creates a secure state parameter containing org ID and provider
func (h *Handler) generateOAuthState(orgID, provider string) (string, error) {
	// Create random bytes for security
	const stateRandomBytesLength = 16
	randomBytes := make([]byte, stateRandomBytesLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Format: orgID:provider:randomBytes (base64 encoded)
	stateData := fmt.Sprintf("%s:%s:%s", orgID, provider, base64.URLEncoding.EncodeToString(randomBytes))
	return base64.URLEncoding.EncodeToString([]byte(stateData)), nil
}

// validateOAuthState validates state parameter and extracts org ID and provider
func (h *Handler) validateOAuthState(state string) (orgID, provider string, err error) {
	// Decode base64
	stateBytes, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return "", "", err
	}

	// Split by colons
	parts := strings.Split(string(stateBytes), ":")
	const expectedStateParts = 3
	if len(parts) != expectedStateParts {
		return "", "", ErrInvalidStateFormat
	}

	return parts[0], parts[1], nil
}

// storeIntegrationTokens creates/updates integration and stores OAuth tokens securely
func (h *Handler) storeIntegrationTokens(ctx context.Context, orgID, provider string, userInfo *IntegrationUserInfo, oauthToken *oauth2.Token) (*ent.Integration, error) {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	systemCtx = contextx.With(systemCtx, auth.OrgSubscriptionContextKey{})

	// Check if integration already exists for this org/provider
	integration, err := h.DBClient.Integration.Query().
		Where(
			integrationschema.And(
				integrationschema.OwnerID(orgID),
				integrationschema.Kind(provider),
			),
		).
		Only(systemCtx)

	if ent.IsNotFound(err) {
		// Create new integration if not found
		integration, err = h.DBClient.Integration.Create().
			SetOwnerID(orgID).
			SetName(fmt.Sprintf("%s Integration", provider)).
			SetDescription(fmt.Sprintf("OAuth integration with %s", provider)).
			SetKind(provider).
			Save(systemCtx)
		if err != nil {
			zerolog.Ctx(systemCtx).Error().Msgf("Failed to create integration for org %s and provider %s: %v", orgID, provider, err)
			return nil, err
		}
	} else if err != nil {
		zerolog.Ctx(systemCtx).Error().Err(err).Msgf("Failed to query integration for org %s and provider %s", orgID, provider)
		return nil, err
	}

	integrationName := fmt.Sprintf("%s Integration", provider)
	if userInfo.Username != "" {
		integrationName = fmt.Sprintf("%s (%s)", integrationName, userInfo.Username)
	}

	// Create or update integration
	if ent.IsNotFound(err) {
		// Create new integration
		integration, err = h.DBClient.Integration.Create().
			SetOwnerID(orgID).
			SetName(integrationName).
			SetDescription(fmt.Sprintf("OAuth integration with %s for %s", provider, userInfo.Username)).
			SetKind(provider).
			Save(systemCtx)
		if err != nil {
			zerolog.Ctx(systemCtx).Error().Msgf("Failed to create integration for org %s and provider %s: %v", orgID, provider, err)
			return nil, fmt.Errorf("failed to create integration: %w", err)
		}
	} else {
		// Update existing integration
		integration, err = integration.Update().
			SetName(integrationName).
			SetDescription(fmt.Sprintf("OAuth integration with %s for %s", provider, userInfo.Username)).
			Save(systemCtx)
		if err != nil {
			zerolog.Ctx(systemCtx).Error().Msgf("Failed to update integration for org %s and provider %s: %v", orgID, provider, err)
			return nil, fmt.Errorf("failed to update integration: %w", err)
		}
	}

	// Store access token
	if err := h.storeSecretForIntegration(systemCtx, integration, "access_token", oauthToken.AccessToken); err != nil {
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	// Store refresh token if available
	if oauthToken.RefreshToken != "" {
		if err := h.storeSecretForIntegration(systemCtx, integration, "refresh_token", oauthToken.RefreshToken); err != nil {
			return nil, fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	// Store token expiry if available
	if !oauthToken.Expiry.IsZero() {
		if err := h.storeSecretForIntegration(systemCtx, integration, "expires_at", oauthToken.Expiry.Format(time.RFC3339)); err != nil {
			return nil, fmt.Errorf("failed to store token expiry: %w", err)
		}
	}

	// Store additional provider-specific metadata
	metadata := map[string]string{
		"provider_user_id":  userInfo.ID,
		"provider_username": userInfo.Username,
	}
	if userInfo.Email != "" {
		metadata["provider_email"] = userInfo.Email
	}

	for key, value := range metadata {
		if err := h.storeSecretForIntegration(systemCtx, integration, key, value); err != nil {
			return nil, fmt.Errorf("failed to store metadata %s: %w", key, err)
		}
	}

	return integration, nil
}

// storeSecretForIntegration creates or updates a hush secret for an integration
func (h *Handler) storeSecretForIntegration(ctx context.Context, integration *ent.Integration, name, value string) error {
	secretName := fmt.Sprintf("%s_%s", integration.Kind, name)

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	// Check if secret already exists
	existing, err := h.DBClient.Hush.Query().
		Where(
			hushschema.And(
				hushschema.OwnerID(integration.OwnerID),
				hushschema.HasIntegrationsWith(integrationschema.ID(integration.ID)),
				hushschema.SecretName(secretName),
			),
		).
		Only(systemCtx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if ent.IsNotFound(err) {
		// Create new secret
		_, err = h.DBClient.Hush.Create().
			SetOwnerID(integration.OwnerID).
			SetName(fmt.Sprintf("%s %s", integration.Name, strings.ReplaceAll(name, "_", " "))).
			SetDescription(fmt.Sprintf("%s for %s integration", cases.Title(language.English).String(strings.ReplaceAll(name, "_", " ")), integration.Kind)).
			SetKind("oauth_token").
			SetSecretName(secretName).
			SetSecretValue(value).
			AddIntegrations(integration).
			Save(systemCtx)
		return err
	}

	// Secret value is immutable, so delete and recreate
	err = h.DBClient.Hush.DeleteOne(existing).Exec(systemCtx)
	if err != nil {
		return fmt.Errorf("failed to delete existing secret: %w", err)
	}

	// Create new secret with updated value
	_, err = h.DBClient.Hush.Create().
		SetOwnerID(integration.OwnerID).
		SetName(fmt.Sprintf("%s %s", integration.Name, strings.ReplaceAll(name, "_", " "))).
		SetDescription(fmt.Sprintf("%s for %s integration", cases.Title(language.English).String(strings.ReplaceAll(name, "_", " ")), integration.Kind)).
		SetKind("oauth_token").
		SetSecretName(secretName).
		SetSecretValue(value).
		AddIntegrations(integration).
		Save(systemCtx)
	return err
}

// validateGithubIntegrationToken validates GitHub token and returns user info
func (h *Handler) validateGithubIntegrationToken(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
	// For integration tokens, we'll need to make a direct API call to GitHub
	// since the github provider's UserFromContext expects a specific flow

	// Create HTTP client with the OAuth token
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub API request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token.AccessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %w", resp.StatusCode, ErrValidateToken)
	}

	// For now, return basic validated info
	// TODO: Parse actual JSON response for complete user data
	return &IntegrationUserInfo{
		ID:       "github_user", // Would extract from JSON response
		Username: "github_user", // Would extract from JSON response
		Email:    "",            // Would extract from JSON response
	}, nil
}

// validateSlackIntegrationToken validates Slack token and returns user info
func (h *Handler) validateSlackIntegrationToken(_ context.Context, _ *oauth2.Token) (*IntegrationUserInfo, error) {
	// Placeholder implementation - would need actual Slack API client
	// For now, return basic info
	return &IntegrationUserInfo{
		ID:       "slack_user_id",
		Username: "slack_user",
		Email:    "",
	}, nil
}

// GetIntegrationToken retrieves stored OAuth tokens for an integration
func (h *Handler) GetIntegrationToken(ctx echo.Context) error {
	provider := ctx.PathParam("provider")
	if provider == "" {
		return h.BadRequest(ctx, ErrProviderRequired)
	}

	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err)
	}

	orgID := user.OrganizationID

	// Get integration token
	tokenData, err := h.retrieveIntegrationToken(userCtx, orgID, provider)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, fmt.Errorf("integration not found for provider %s: %w", provider, ErrIntegrationNotFound))
		}
		return h.InternalServerError(ctx, err)
	}

	out := models.IntegrationTokenResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  provider,
		Token:     tokenData,
		ExpiresAt: tokenData.ExpiresAt,
	}

	return h.Success(ctx, out)
}

// ListIntegrations returns all integrations for the user's organization
func (h *Handler) ListIntegrations(ctx echo.Context) error {
	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err)
	}

	orgID := user.OrganizationID

	// Query integrations for the organization
	integrations, err := h.DBClient.Integration.Query().
		Where(integrationschema.OwnerID(orgID)).
		All(userCtx)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	out := models.ListIntegrationsResponse{
		Reply:        rout.Reply{Success: true},
		Integrations: integrations,
	}

	return h.Success(ctx, out)
}

// DeleteIntegration removes an integration and its associated secrets
func (h *Handler) DeleteIntegration(ctx echo.Context) error {
	integrationID := ctx.PathParam("id")
	if integrationID == "" {
		return h.BadRequest(ctx, ErrIntegrationIDRequired)
	}

	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err)
	}

	orgID := user.OrganizationID

	// Verify integration belongs to user's organization
	integration, err := h.DBClient.Integration.Query().
		Where(
			integrationschema.And(
				integrationschema.ID(integrationID),
				integrationschema.OwnerID(orgID),
			),
		).
		Only(userCtx)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, ErrIntegrationNotFound)
		}
		return h.InternalServerError(ctx, err)
	}

	// Use privacy bypass for deletion
	systemCtx := privacy.DecisionContext(userCtx, privacy.Allow)

	// Delete associated secrets first
	_, err = h.DBClient.Hush.Delete().
		Where(
			hushschema.And(
				hushschema.OwnerID(orgID),
				hushschema.HasIntegrationsWith(integrationschema.ID(integrationID)),
			),
		).
		Exec(systemCtx)
	if err != nil {
		log.Error().Err(err).Str("integration_id", integrationID).Msg("failed to delete integration secrets")
		return h.InternalServerError(ctx, ErrDeleteSecrets)
	}

	// Delete integration
	err = h.DBClient.Integration.DeleteOneID(integrationID).Exec(systemCtx)
	if err != nil {
		return h.InternalServerError(ctx, err)
	}

	out := models.DeleteIntegrationResponse{
		Reply:   rout.Reply{Success: true},
		Message: fmt.Sprintf("Integration %s deleted successfully", integration.Name),
	}

	return h.Success(ctx, out)
}

// retrieveIntegrationToken gets stored OAuth token for a provider
func (h *Handler) retrieveIntegrationToken(ctx context.Context, orgID, provider string) (*models.IntegrationToken, error) {
	// Get integration
	integration, err := h.DBClient.Integration.Query().
		Where(
			integrationschema.And(
				integrationschema.OwnerID(orgID),
				integrationschema.Kind(provider),
			),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// Get associated secrets
	secrets, err := h.DBClient.Hush.Query().
		Where(
			hushschema.And(
				hushschema.OwnerID(orgID),
				hushschema.HasIntegrationsWith(integrationschema.ID(integration.ID)),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Build token data
	tokenData := &models.IntegrationToken{
		Provider: provider,
	}

	for _, secret := range secrets {
		if secret.SecretName == "" || secret.SecretValue == "" {
			continue
		}

		switch secret.SecretName {
		case fmt.Sprintf("%s_access_token", provider):
			tokenData.AccessToken = secret.SecretValue
		case fmt.Sprintf("%s_refresh_token", provider):
			tokenData.RefreshToken = secret.SecretValue
		case fmt.Sprintf("%s_expires_at", provider):
			if expiresAt, err := time.Parse(time.RFC3339, secret.SecretValue); err == nil {
				tokenData.ExpiresAt = &expiresAt
			}
		case fmt.Sprintf("%s_provider_user_id", provider):
			tokenData.ProviderUserID = secret.SecretValue
		case fmt.Sprintf("%s_provider_username", provider):
			tokenData.ProviderUsername = secret.SecretValue
		case fmt.Sprintf("%s_provider_email", provider):
			tokenData.ProviderEmail = secret.SecretValue
		}
	}

	if tokenData.AccessToken == "" {
		return nil, fmt.Errorf("no access token found for provider %s: %w", provider, ErrIntegrationNotFound)
	}

	return tokenData, nil
}

// RefreshIntegrationToken refreshes an expired OAuth token if refresh token is available
func (h *Handler) RefreshIntegrationToken(ctx context.Context, orgID, provider string) (*models.IntegrationToken, error) {
	// Get current token data
	tokenData, err := h.retrieveIntegrationToken(ctx, orgID, provider)
	if err != nil {
		return nil, err
	}

	if tokenData.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available for provider %s: %w", provider, ErrIntegrationNotFound)
	}

	// Get provider configuration
	providers := h.getIntegrationProviders()
	providerConfig, exists := providers[provider]
	if !exists {
		return nil, ErrInvalidProvider
	}

	// Create token source for refresh
	token := &oauth2.Token{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
	}
	if tokenData.ExpiresAt != nil {
		token.Expiry = *tokenData.ExpiresAt
	}

	// Use token source to get a fresh token
	tokenSource := providerConfig.Config.TokenSource(ctx, token)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Validate fresh token
	_, err = providerConfig.Validate(ctx, freshToken)
	if err != nil {
		return nil, fmt.Errorf("refreshed token validation failed: %w", err)
	}

	// Update stored tokens
	integration, err := h.DBClient.Integration.Query().
		Where(
			integrationschema.And(
				integrationschema.OwnerID(orgID),
				integrationschema.Kind(provider),
			),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Update tokens
	if err := h.storeSecretForIntegration(systemCtx, integration, "access_token", freshToken.AccessToken); err != nil {
		return nil, err
	}

	if freshToken.RefreshToken != "" && freshToken.RefreshToken != tokenData.RefreshToken {
		if err := h.storeSecretForIntegration(systemCtx, integration, "refresh_token", freshToken.RefreshToken); err != nil {
			return nil, err
		}
	}

	if !freshToken.Expiry.IsZero() {
		if err := h.storeSecretForIntegration(systemCtx, integration, "expires_at", freshToken.Expiry.Format(time.RFC3339)); err != nil {
			return nil, err
		}
	}

	// Return updated token data
	return h.retrieveIntegrationToken(ctx, orgID, provider)
}

// RefreshIntegrationTokenHandler is the HTTP handler for refreshing integration tokens
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context) error {
	provider := ctx.PathParam("provider")
	if provider == "" {
		return h.BadRequest(ctx, ErrProviderRequired)
	}

	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err)
	}

	orgID := user.OrganizationID

	// Refresh the token
	tokenData, err := h.RefreshIntegrationToken(userCtx, orgID, provider)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, fmt.Errorf("integration not found for provider %s: %w", provider, ErrIntegrationNotFound))
		}
		return h.InternalServerError(ctx, fmt.Errorf("failed to refresh token: %w", err))
	}

	out := models.IntegrationTokenResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  provider,
		Token:     tokenData,
		ExpiresAt: tokenData.ExpiresAt,
	}

	return h.Success(ctx, out)
}

// GetIntegrationStatus checks if an integration is connected and returns its status
func (h *Handler) GetIntegrationStatus(ctx echo.Context) error {
	provider := ctx.PathParam("provider")
	if provider == "" {
		return h.BadRequest(ctx, ErrProviderRequired)
	}

	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()
	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err)
	}

	orgID := user.OrganizationID

	// Check if integration exists
	integration, err := h.DBClient.Integration.Query().
		Where(
			integrationschema.And(
				integrationschema.OwnerID(orgID),
				integrationschema.Kind(provider),
			),
		).
		Only(userCtx)

	if err != nil {
		if ent.IsNotFound(err) {
			out := models.IntegrationStatusResponse{
				Reply:     rout.Reply{Success: true},
				Provider:  provider,
				Connected: false,
				Message:   fmt.Sprintf("No %s integration found", provider),
			}
			return h.Success(ctx, out)
		}
		return h.InternalServerError(ctx, err)
	}

	// Check if we have valid tokens
	tokenData, err := h.retrieveIntegrationToken(userCtx, orgID, provider)
	tokenValid := err == nil && tokenData.AccessToken != ""

	// Check if token is expired
	tokenExpired := false
	if tokenData != nil && tokenData.ExpiresAt != nil {
		tokenExpired = time.Now().After(*tokenData.ExpiresAt)
	}

	status := "connected"
	message := fmt.Sprintf("%s integration is connected and active", cases.Title(language.English).String(provider))

	if !tokenValid {
		status = "invalid"
		message = fmt.Sprintf("%s integration exists but has invalid tokens", cases.Title(language.English).String(provider))
	} else if tokenExpired {
		status = "expired"
		message = fmt.Sprintf("%s integration tokens have expired", cases.Title(language.English).String(provider))
	}

	out := models.IntegrationStatusResponse{
		Reply:        rout.Reply{Success: true},
		Provider:     provider,
		Connected:    true,
		Status:       status,
		TokenValid:   tokenValid,
		TokenExpired: tokenExpired,
		Message:      message,
		Integration:  integration,
	}

	return h.Success(ctx, out)
}

// BindStartOAuthFlowHandler binds the start OAuth flow handler to the OpenAPI schema
func (h *Handler) BindStartOAuthFlowHandler() *openapi3.Operation {
	startOAuthHandler := openapi3.NewOperation()
	startOAuthHandler.Description = "Start OAuth integration flow for third-party providers"
	startOAuthHandler.Tags = []string{"oauth", "integrations"}
	startOAuthHandler.OperationID = "StartOAuthFlow"
	startOAuthHandler.Security = AllSecurityRequirements()
	h.AddRequestBody("OAuthFlowRequest", models.ExampleOAuthFlowRequest, startOAuthHandler)
	h.AddResponse("OAuthFlowResponse", "success", models.ExampleOAuthFlowResponse, startOAuthHandler, http.StatusOK)
	startOAuthHandler.AddResponse(http.StatusInternalServerError, internalServerError())
	startOAuthHandler.AddResponse(http.StatusBadRequest, badRequest())
	startOAuthHandler.AddResponse(http.StatusUnauthorized, unauthorized())

	return startOAuthHandler
}

// BindHandleOAuthCallbackHandler binds the OAuth callback handler to the OpenAPI schema
func (h *Handler) BindHandleOAuthCallbackHandler() *openapi3.Operation {
	callbackHandler := openapi3.NewOperation()
	callbackHandler.Description = "Handle OAuth callback and store integration tokens"
	callbackHandler.Tags = []string{"oauth", "integrations"}
	callbackHandler.OperationID = "HandleOAuthCallback"
	callbackHandler.Security = AllSecurityRequirements()
	h.AddRequestBody("OAuthCallbackRequest", models.ExampleOAuthCallbackRequest, callbackHandler)
	h.AddResponse("OAuthCallbackResponse", "success", models.ExampleOAuthCallbackResponse, callbackHandler, http.StatusOK)
	callbackHandler.AddResponse(http.StatusInternalServerError, internalServerError())
	callbackHandler.AddResponse(http.StatusBadRequest, badRequest())
	callbackHandler.AddResponse(http.StatusUnauthorized, unauthorized())

	return callbackHandler
}
