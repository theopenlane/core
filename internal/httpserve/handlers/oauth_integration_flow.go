package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/theopenlane/httpsling"
	"golang.org/x/oauth2"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	ent "github.com/theopenlane/core/internal/ent/generated"
	hushschema "github.com/theopenlane/core/internal/ent/generated/hush"
	integrationschema "github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/logx"
	models "github.com/theopenlane/core/pkg/openapi"
)

var errSlackAPI = errors.New("slack API error")

// OAuth error helpers
func wrapAPIError(provider string, err error) error {
	return fmt.Errorf("failed to call %s API: %w", provider, err)
}

func wrapIntegrationError(operation string, err error) error {
	return fmt.Errorf("failed to %s integration: %w", operation, err)
}

func wrapSecretError(operation, secretType string, err error) error {
	return fmt.Errorf("failed to %s %s: %w", operation, secretType, err)
}

func wrapTokenError(operation, provider string, err error) error {
	return fmt.Errorf("failed to %s token for %s: %w", operation, provider, err)
}

var (
	oauthStateCookieName  = "oauth_state"
	oauthOrgIDCookieName  = "oauth_org_id"
	oauthUserIDCookieName = "oauth_user_id"
)

// StartOAuthFlow initiates the OAuth flow for a third-party integration
func (h *Handler) StartOAuthFlow(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleOAuthFlowRequest, models.ExampleOAuthFlowResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	userCtx := ctx.Request().Context()
	respWrite := ctx.Response().Writer

	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	// Get user's current organization
	orgID := user.OrganizationID

	// Validate provider
	providers := h.getIntegrationProviders()

	provider, exists := providers[strings.ToLower(in.Provider)]
	if !exists {
		return h.BadRequest(ctx, ErrInvalidProvider, openapi)
	}

	// Set up cookie config that will work in either prod, test, or development mode
	cfg := h.getOauthCookieConfig()

	// Set the org ID and user ID as cookies for the OAuth flow
	sessions.SetCookie(respWrite, orgID, oauthOrgIDCookieName, cfg)
	sessions.SetCookie(respWrite, user.SubjectID, oauthUserIDCookieName, cfg)

	// Re-set existing auth cookies with SameSiteNone for OAuth compatibility
	if accessCookie, err := sessions.GetCookie(ctx.Request(), auth.AccessTokenCookie); err == nil {
		sessions.SetCookie(respWrite, accessCookie.Value, auth.AccessTokenCookie, cfg)
	}

	if refreshCookie, err := sessions.GetCookie(ctx.Request(), auth.RefreshTokenCookie); err == nil {
		sessions.SetCookie(respWrite, refreshCookie.Value, auth.RefreshTokenCookie, cfg)
	}

	// Generate state parameter for security
	state, err := h.generateOAuthState(orgID, in.Provider)
	if err != nil {
		logx.FromContext(userCtx).Error().Err(err).Msg("failed to generate oauth state")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// Set the state as a cookie for validation in callback
	sessions.SetCookie(respWrite, state, oauthStateCookieName, cfg)

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
func (h *Handler) HandleOAuthCallback(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.OAuthCallbackRequest{}, models.ExampleOAuthCallbackResponse, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Skip actual handler logic during OpenAPI registration
	if isRegistrationContext(ctx) {
		return nil
	}

	reqCtx := ctx.Request().Context()

	// Validate state matches what was set in the cookie
	stateCookie, err := sessions.GetCookie(ctx.Request(), oauthStateCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_state cookie")
		return h.BadRequest(ctx, ErrInvalidState, openapi)
	}

	if in.State != stateCookie.Value {
		logx.FromContext(reqCtx).Error().Str("payload state", in.State).Str("cookie state", stateCookie.Value).Msg("State cookies do not match")

		return h.BadRequest(ctx, ErrInvalidState, openapi)
	}

	// Get org ID and user ID from cookies
	orgCookie, err := sessions.GetCookie(ctx.Request(), oauthOrgIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_org_id cookie")

		return h.BadRequest(ctx, ErrMissingOrganizationContext, openapi)
	}

	orgID := orgCookie.Value

	_, err = sessions.GetCookie(ctx.Request(), oauthUserIDCookieName)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to get oauth_user_id cookie")

		return h.BadRequest(ctx, ErrMissingUserContext, openapi)
	}

	// Get the user from database to set authenticated context
	respWrite := ctx.Response().Writer

	systemCtx := privacy.DecisionContext(reqCtx, privacy.Allow)

	// Validate state and extract provider from state
	_, provider, err := h.validateOAuthState(in.State)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("failed to validate oauth state")

		return h.BadRequest(ctx, ErrInvalidState, openapi)
	}

	// Set the provider from the state (GitHub doesn't send provider in callback)
	in.Provider = provider

	// Get provider configuration
	providers := h.getIntegrationProviders()

	providerConfig, exists := providers[provider]
	if !exists {
		return h.BadRequest(ctx, ErrInvalidProvider, openapi)
	}

	// Exchange code for token
	oauthToken, err := providerConfig.Config.Exchange(reqCtx, in.Code)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("provider", provider).Msg("failed to exchange OAuth code for token")

		return h.InternalServerError(ctx, ErrExchangeAuthCode, openapi)
	}

	// Validate token and get user info
	userInfo, err := providerConfig.Validate(reqCtx, oauthToken)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("provider", provider).Msg("failed to validate OAuth token")

		return h.InternalServerError(ctx, ErrValidateToken, openapi)
	}

	// Store integration and tokens (use authenticated context)
	_, err = h.storeIntegrationTokens(systemCtx, orgID, provider, userInfo, oauthToken)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("provider", provider).Str("org_id", orgID).Msg("failed to store integration tokens")

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	// Clean up the OAuth cookies after successful completion (like SSO handler)
	sessions.RemoveCookie(respWrite, oauthStateCookieName, sessions.CookieConfig{Path: "/"})
	sessions.RemoveCookie(respWrite, oauthOrgIDCookieName, sessions.CookieConfig{Path: "/"})
	sessions.RemoveCookie(respWrite, oauthUserIDCookieName, sessions.CookieConfig{Path: "/"})

	// Redirect to configured success URL with integration details
	helper := NewIntegrationHelper(provider, "")
	redirectURL := helper.RedirectURL(h.IntegrationOauthProvider.SuccessRedirectURL)

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
	helper := NewIntegrationHelper(provider, "")
	stateData := helper.StateData(orgID, randomBytes)

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
	integration, err := h.DBClient.Integration.Query().Where(integrationschema.And(integrationschema.OwnerID(orgID), integrationschema.Kind(provider))).First(systemCtx)

	if !ent.IsNotFound(err) && err != nil {
		return nil, err
	}

	helperWithUser := NewIntegrationHelper(provider, userInfo.Username)
	integrationName := helperWithUser.Name()

	if ent.IsNotFound(err) {
		// Create new integration
		integration, err = h.DBClient.Integration.Create().SetOwnerID(orgID).SetName(integrationName).SetDescription(helperWithUser.Description()).SetKind(provider).Save(systemCtx)
		if err != nil {
			logx.FromContext(systemCtx).Error().Msgf("failed to create integration for org %s and provider %s: %v", orgID, provider, err)

			return nil, wrapIntegrationError("create", err)
		}
	} else {
		// Update existing integration
		integration, err = integration.Update().SetName(integrationName).SetDescription(helperWithUser.Description()).Save(systemCtx)
		if err != nil {
			logx.FromContext(systemCtx).Error().Msgf("failed to update integration for org %s and provider %s: %v", orgID, provider, err)

			return nil, wrapIntegrationError("update", err)
		}
	}

	// Store access token
	if err := h.storeSecretForIntegration(systemCtx, integration, "access_token", oauthToken.AccessToken); err != nil {
		return nil, wrapSecretError("store", "access token", err)
	}

	// Store refresh token if available
	if oauthToken.RefreshToken != "" {
		if err := h.storeSecretForIntegration(systemCtx, integration, "refresh_token", oauthToken.RefreshToken); err != nil {
			return nil, wrapSecretError("store", "refresh token", err)
		}
	}

	// Store token expiry if available
	if !oauthToken.Expiry.IsZero() {
		if err := h.storeSecretForIntegration(systemCtx, integration, "expires_at", oauthToken.Expiry.Format(time.RFC3339)); err != nil {
			return nil, wrapSecretError("store", "token expiry", err)
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
			return nil, wrapSecretError("store", fmt.Sprintf("metadata %s", key), err)
		}
	}

	return integration, nil
}

// storeSecretForIntegration creates or updates a hush secret for an integration
func (h *Handler) storeSecretForIntegration(ctx context.Context, integration *ent.Integration, name, value string) error {
	helper := NewIntegrationHelper(integration.Kind, "")
	secretName := helper.SecretName(name)
	description := helper.SecretDescription(name)

	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)
	// Check if secret already exists
	existing, err := h.DBClient.Hush.Query().Where(hushschema.And(hushschema.OwnerID(integration.OwnerID),
		hushschema.HasIntegrationsWith(integrationschema.ID(integration.ID)),
		hushschema.SecretName(secretName))).
		Only(systemCtx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if ent.IsNotFound(err) {
		// Create new secret
		if err := h.DBClient.Hush.Create().SetOwnerID(integration.OwnerID).
			SetName(helper.SecretDisplayName(integration.Name, name)).
			SetDescription(description).
			SetKind(oauthTokenKind).
			SetSecretName(secretName).
			SetSecretValue(value).
			AddIntegrations(integration).
			Exec(systemCtx); err != nil {
			return err
		}

		return nil
	}

	// Secret value is immutable, so delete and recreate
	err = h.DBClient.Hush.DeleteOne(existing).Exec(systemCtx)
	if err != nil {
		return wrapSecretError("delete", "existing secret", err)
	}

	// Create new secret with updated value
	if err := h.DBClient.Hush.Create().
		SetOwnerID(integration.OwnerID).
		SetName(helper.SecretDisplayName(integration.Name, name)).
		SetDescription(description).
		SetKind(oauthTokenKind).
		SetSecretName(secretName).
		SetSecretValue(value).
		AddIntegrations(integration).
		Exec(systemCtx); err != nil {
		return err
	}

	return nil
}

// validateGithubIntegrationToken validates GitHub token and returns user info
func (h *Handler) validateGithubIntegrationToken(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
	// For integration tokens, we'll need to make a direct API call to GitHub
	// since the github provider's UserFromContext expects a specific flow
	helper := NewIntegrationHelper("github", "")
	headerName, headerValue := helper.AuthHeader(token.AccessToken)

	// Use httpsling to make the API call
	var githubUser GitHubUser

	resp, err := httpsling.ReceiveWithContext(ctx, &githubUser,
		httpsling.Get("https://api.github.com/user"),
		httpsling.Header(headerName, headerValue),
		httpsling.Header("Accept", "application/vnd.github.v3+json"),
		httpsling.ExpectCode(http.StatusOK),
	)
	if err != nil {
		return nil, wrapAPIError("GitHub", err)
	}

	defer resp.Body.Close()

	// Convert GitHub user to IntegrationUserInfo
	userInfo := &IntegrationUserInfo{
		ID:       strconv.Itoa(githubUser.ID),
		Username: githubUser.Login,
		Email:    githubUser.Email,
	}

	// GitHub's primary email might not be public, so we need to fetch it separately if empty
	if userInfo.Email == "" {
		if email, err := h.getGithubUserEmail(ctx, token.AccessToken); err == nil {
			userInfo.Email = email
		}
	}

	return userInfo, nil
}

// getGithubUserEmail fetches user's email from GitHub emails API
func (h *Handler) getGithubUserEmail(ctx context.Context, accessToken string) (string, error) {
	helper := NewIntegrationHelper("github", "")
	headerName, headerValue := helper.AuthHeader(accessToken)

	// Use httpsling to make the API call
	var emails []GitHubEmail

	resp, err := httpsling.ReceiveWithContext(ctx, &emails,
		httpsling.Get("https://api.github.com/user/emails"),
		httpsling.Header(headerName, headerValue),
		httpsling.Header("Accept", "application/vnd.github.v3+json"),
		httpsling.ExpectCode(http.StatusOK),
	)
	if err != nil {
		return "", wrapAPIError("GitHub emails", err)
	}

	defer resp.Body.Close()

	// Find the primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// If no primary verified email, find any verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	// If no verified emails, return the first one
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", ErrIntegrationNotFound
}

// validateSlackIntegrationToken validates Slack token and returns user info
func (h *Handler) validateSlackIntegrationToken(ctx context.Context, token *oauth2.Token) (*IntegrationUserInfo, error) {
	helper := NewIntegrationHelper("slack", "")
	headerName, headerValue := helper.AuthHeader(token.AccessToken)

	// Use httpsling to make the API call
	var slackResp SlackUser

	resp, err := httpsling.ReceiveWithContext(ctx, &slackResp,
		httpsling.Get("https://slack.com/api/users.identity"),
		httpsling.Header(headerName, headerValue),
		httpsling.Header("Content-Type", "application/x-www-form-urlencoded"),
		httpsling.ExpectCode(http.StatusOK),
	)
	if err != nil {
		return nil, wrapAPIError("Slack", err)
	}

	defer resp.Body.Close()

	// Check if the response is successful
	if !slackResp.OK {
		return nil, fmt.Errorf("%w: %s", errSlackAPI, slackResp.Error)
	}

	// Convert Slack user to IntegrationUserInfo
	userInfo := &IntegrationUserInfo{
		ID:       slackResp.User.ID,
		Username: slackResp.User.Name,
		Email:    slackResp.User.Profile.Email,
	}

	// Use display name if username is empty
	if userInfo.Username == "" && slackResp.User.Profile.DisplayName != "" {
		userInfo.Username = slackResp.User.Profile.DisplayName
	}

	return userInfo, nil
}

// retrieveIntegrationToken gets stored OAuth token for a provider
func (h *Handler) retrieveIntegrationToken(ctx context.Context, orgID, provider string) (*models.IntegrationToken, error) {
	// Get integration with associated secrets in a single query
	integration, err := h.DBClient.Integration.Query().
		Where(integrationschema.And(integrationschema.OwnerID(orgID), integrationschema.Kind(provider))).
		WithSecrets().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// Build token data from the loaded secrets
	tokenData := &models.IntegrationToken{
		Provider: provider,
	}

	for _, secret := range integration.Edges.Secrets {
		if secret.SecretName == "" || secret.SecretValue == "" {
			continue
		}

		switch secret.SecretName {
		case provider + secretNameSeparator + accessTokenField:
			tokenData.AccessToken = secret.SecretValue
		case provider + secretNameSeparator + refreshTokenField:
			tokenData.RefreshToken = secret.SecretValue
		case provider + secretNameSeparator + expiresAtField:
			if expiresAt, err := time.Parse(time.RFC3339, secret.SecretValue); err == nil {
				tokenData.ExpiresAt = &expiresAt
			}
		case provider + secretNameSeparator + providerUserIDField:
			tokenData.ProviderUserID = secret.SecretValue
		case provider + secretNameSeparator + providerUsernameField:
			tokenData.ProviderUsername = secret.SecretValue
		case provider + secretNameSeparator + providerEmailField:
			tokenData.ProviderEmail = secret.SecretValue
		}
	}

	if tokenData.AccessToken == "" {
		return nil, wrapTokenError("find access", provider, ErrIntegrationNotFound)
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
		return nil, wrapTokenError("find refresh", provider, ErrIntegrationNotFound)
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
		return nil, wrapTokenError("refresh", provider, err)
	}

	// Validate fresh token
	_, err = providerConfig.Validate(ctx, freshToken)
	if err != nil {
		return nil, wrapTokenError("validate refreshed", provider, err)
	}

	// Update stored tokens
	integration, err := h.DBClient.Integration.Query().Where(integrationschema.And(integrationschema.OwnerID(orgID), integrationschema.Kind(provider))).First(ctx)
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
func (h *Handler) RefreshIntegrationTokenHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleRefreshIntegrationTokenRequest, models.IntegrationTokenResponse{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	// Get the authenticated user and organization
	userCtx := ctx.Request().Context()

	user, err := auth.GetAuthenticatedUserFromContext(userCtx)
	if err != nil {
		return h.Unauthorized(ctx, err, openapi)
	}

	orgID := user.OrganizationID

	// Refresh the token
	tokenData, err := h.RefreshIntegrationToken(userCtx, orgID, in.Provider)
	if err != nil {
		if ent.IsNotFound(err) {
			return h.NotFound(ctx, wrapIntegrationError("find", fmt.Errorf("provider %s: %w", in.Provider, ErrIntegrationNotFound)))
		}

		return h.InternalServerError(ctx, wrapTokenError("refresh", in.Provider, err), openapi)
	}

	out := models.IntegrationTokenResponse{
		Reply:     rout.Reply{Success: true},
		Provider:  in.Provider,
		Token:     tokenData,
		ExpiresAt: tokenData.ExpiresAt,
	}

	return h.Success(ctx, out)
}

// getOauthCookieConfig returns the cookie configuration for OAuth cookies
// that is dependent on if the environment is test/dev or production
func (h *Handler) getOauthCookieConfig() sessions.CookieConfig {
	secure := !h.IsTest && !h.IsDev

	sameSite := http.SameSiteNoneMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}

	cfg := sessions.CookieConfig{
		Path:     "/",
		HTTPOnly: true,
		Secure:   secure, // Must be true with SameSiteNone in production
		SameSite: sameSite,
	}

	return cfg
}
