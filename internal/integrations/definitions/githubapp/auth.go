package githubapp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gh "github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultJWTExpiry           = 9 * time.Minute
	jwtIssuedAtBackdateSeconds = 30 * time.Second
)

// Auth executes the GitHub App install flow
type Auth struct {
	// Config holds the operator-supplied GitHub App settings
	Config Config
}

// InstallStartInput carries runtime state used to construct the install URL
type InstallStartInput struct {
	// State carries caller-provided state through the install redirect
	State string `json:"state"`
}

// InstallationCallbackState carries the installation ID received from the GitHub callback
type InstallationCallbackState struct {
	// InstallationID is the GitHub installation selected during the callback flow
	InstallationID string `json:"installationId"`
}

// ProviderData is the provider-specific data stored in the credential ProviderData field
type ProviderData struct {
	// AppID is the GitHub App identifier used to mint installation tokens
	AppID string `json:"appId"`

	// InstallationID is the installation selected for this credential
	InstallationID string `json:"installationId"`
}

// Start begins the GitHub App install flow by returning the installation URL
func (a Auth) Start(ctx context.Context, input json.RawMessage) (types.AuthStartResult, error) {
	if err := ctx.Err(); err != nil {
		return types.AuthStartResult{}, err
	}

	var startInput InstallStartInput
	if err := jsonx.UnmarshalIfPresent(input, &startInput); err != nil {
		return types.AuthStartResult{}, err
	}

	if a.Config.AppSlug == "" {
		return types.AuthStartResult{}, ErrAppSlugMissing
	}

	installURL := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/apps/" + a.Config.AppSlug + "/installations/new",
	}
	if startInput.State != "" {
		query := installURL.Query()
		query.Set("state", startInput.State)
		installURL.RawQuery = query.Encode()
	}

	return types.AuthStartResult{URL: installURL.String()}, nil
}

// Complete finishes the GitHub App install flow by minting an installation access token
func (a Auth) Complete(ctx context.Context, _ json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	if err := ctx.Err(); err != nil {
		return types.AuthCompleteResult{}, err
	}

	var state InstallationCallbackState
	if err := jsonx.UnmarshalIfPresent(input, &state); err != nil {
		return types.AuthCompleteResult{}, err
	}

	if state.InstallationID == "" {
		return types.AuthCompleteResult{}, ErrInstallationIDMissing
	}

	if a.Config.AppID == "" {
		return types.AuthCompleteResult{}, ErrAppIDMissing
	}

	privateKey := normalizePrivateKey(a.Config.PrivateKey)
	if privateKey == "" {
		return types.AuthCompleteResult{}, ErrPrivateKeyMissing
	}

	jwtToken, err := a.appJWT(privateKey)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	installationToken, err := a.installationToken(ctx, state.InstallationID, jwtToken)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	providerData, err := jsonx.ToRawMessage(ProviderData{
		AppID:          a.Config.AppID,
		InstallationID: state.InstallationID,
	})
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	credential := types.CredentialSet{
		OAuthAccessToken:  installationToken.AccessToken,
		OAuthRefreshToken: installationToken.RefreshToken,
		OAuthTokenType:    installationToken.TokenType,
		ProviderData:      providerData,
	}

	if !installationToken.Expiry.IsZero() {
		expiry := installationToken.Expiry.UTC()
		credential.OAuthExpiry = &expiry
	}

	return types.AuthCompleteResult{Credential: credential}, nil
}

// appJWT signs a short-lived JWT for GitHub App authentication
func (a Auth) appJWT(privateKey string) (string, error) {
	key, err := parseRSAPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    a.Config.AppID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-jwtIssuedAtBackdateSeconds)),
		ExpiresAt: jwt.NewNumericDate(now.Add(defaultJWTExpiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", ErrJWTSigningFailed
	}

	return signed, nil
}

// parseRSAPrivateKey parses PKCS#1 or PKCS#8 encoded RSA private keys
func parseRSAPrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, ErrPrivateKeyInvalid
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, ErrPrivateKeyInvalid
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrPrivateKeyInvalid
	}

	return rsaKey, nil
}

// installationToken exchanges an app JWT for an installation access token
func (a Auth) installationToken(ctx context.Context, installationID string, jwtToken string) (*oauth2.Token, error) {
	if installationID == "" {
		return nil, ErrInstallationIDMissing
	}

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, ErrInstallationTokenRequestFailed
	}

	client := a.installationTokenClient(ctx, jwtToken)
	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationIDInt, &gh.InstallationTokenOptions{})
	if err != nil {
		return nil, ErrInstallationTokenRequestFailed
	}

	if installationToken.GetToken() == "" {
		return nil, ErrInstallationTokenEmpty
	}

	token := &oauth2.Token{
		AccessToken: installationToken.GetToken(),
		TokenType:   "Bearer",
	}

	expiresAt := installationToken.GetExpiresAt().Time
	if !expiresAt.IsZero() {
		token.Expiry = expiresAt
	}

	return token, nil
}

// installationTokenClient builds the GitHub API client used for installation token requests
func (a Auth) installationTokenClient(ctx context.Context, jwtToken string) *gh.Client {
	source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken})
	httpClient := oauth2.NewClient(ctx, source)
	client := gh.NewClient(httpClient)

	baseURL := strings.TrimRight(a.Config.BaseURL, "/")
	if baseURL == "" || baseURL == githubAPIBaseURL {
		return client
	}

	uploadURL := strings.TrimSuffix(baseURL, "/api/v3")
	if uploadURL == "" {
		uploadURL = baseURL
	}

	enterpriseClient, err := client.WithEnterpriseURLs(baseURL, uploadURL)
	if err != nil {
		return client
	}

	return enterpriseClient
}

// normalizePrivateKey rewrites escaped newlines into PEM newlines when needed
func normalizePrivateKey(value string) string {
	if strings.Contains(value, "\\n") && !strings.Contains(value, "\n") {
		return strings.ReplaceAll(value, "\\n", "\n")
	}

	return value
}
