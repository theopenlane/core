package githubapp

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
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

// installationCallbackState carries the installation ID received from the GitHub callback
type installationCallbackState struct {
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId"`
}

// appProviderData is the provider-specific data stored in the credential ProviderData field
type appProviderData struct {
	// AppID is the GitHub App identifier
	AppID string `json:"appId"`
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId"`
}

// startInstallAuth starts the GitHub App install flow by returning the installation URL
func (d *def) startInstallAuth(_ context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	if d.cfg.AppSlug == "" {
		return types.AuthStartResult{}, ErrAppSlugMissing
	}

	return types.AuthStartResult{
		URL: fmt.Sprintf("https://github.com/apps/%s/installations/new", d.cfg.AppSlug),
	}, nil
}

// completeInstallAuth completes the GitHub App install flow by minting an installation access token
func (d *def) completeInstallAuth(ctx context.Context, _ json.RawMessage, input json.RawMessage) (types.AuthCompleteResult, error) {
	var state installationCallbackState
	if err := jsonx.UnmarshalIfPresent(input, &state); err != nil {
		return types.AuthCompleteResult{}, err
	}

	if state.InstallationID == "" {
		return types.AuthCompleteResult{}, ErrInstallationIDMissing
	}

	if d.cfg.AppID == "" {
		return types.AuthCompleteResult{}, ErrAppIDMissing
	}

	privateKey := normalizePrivateKey(d.cfg.PrivateKey)
	if privateKey == "" {
		return types.AuthCompleteResult{}, ErrPrivateKeyMissing
	}

	jwtToken, err := buildAppJWT(d.cfg.AppID, privateKey)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	baseURL := strings.TrimRight(d.cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = githubAPIBaseURL
	}

	installationToken, err := requestInstallationToken(ctx, state.InstallationID, jwtToken, baseURL)
	if err != nil {
		return types.AuthCompleteResult{}, err
	}

	providerData, err := jsonx.ToRawMessage(appProviderData{
		AppID:          d.cfg.AppID,
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
		exp := installationToken.Expiry.UTC()
		credential.OAuthExpiry = &exp
	}

	return types.AuthCompleteResult{Credential: credential}, nil
}

// buildAppJWT signs a GitHub App JWT for authentication
func buildAppJWT(appID, privateKey string) (string, error) {
	key, err := parseRSAPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    appID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-jwtIssuedAtBackdateSeconds)),
		ExpiresAt: jwt.NewNumericDate(now.Add(defaultJWTExpiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrJWTSigningFailed, err)
	}

	return signed, nil
}

// parseRSAPrivateKey parses a PEM-encoded RSA private key
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

// requestInstallationToken exchanges a GitHub App JWT for an installation access token
func requestInstallationToken(ctx context.Context, installationID, jwtToken, baseURL string) (*oauth2.Token, error) {
	if installationID == "" {
		return nil, ErrInstallationIDMissing
	}

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInstallationTokenRequestFailed, err)
	}

	client, err := newGitHubAPIClient(ctx, jwtToken, baseURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInstallationTokenRequestFailed, err)
	}

	installationToken, _, err := client.Apps.CreateInstallationToken(ctx, installationIDInt, &gh.InstallationTokenOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInstallationTokenRequestFailed, err)
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

// normalizePrivateKey converts escaped newlines to PEM newlines
func normalizePrivateKey(value string) string {
	if strings.Contains(value, "\\n") && !strings.Contains(value, "\n") {
		return strings.ReplaceAll(value, "\\n", "\n")
	}

	return value
}
