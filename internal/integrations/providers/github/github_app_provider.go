package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
)

const (
	// TypeGitHubApp identifies the GitHub App provider
	TypeGitHubApp = types.ProviderType("github_app")
)

// Default GitHub App API configuration values.
const (
	defaultGitHubAPIBaseURL = "https://api.github.com"
	defaultJWTExpiry        = 9 * time.Minute
	defaultHTTPTimeout      = 10 * time.Second
)

// AppBuilder returns the GitHub App provider builder.
func AppBuilder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGitHubApp,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.AuthType != "" && spec.AuthType != types.AuthKindGitHubApp {
				return nil, fmt.Errorf("%w (provider %s expects %s, found %s)", ErrAuthTypeMismatch, TypeGitHubApp, types.AuthKindGitHubApp, spec.AuthType)
			}

			baseURL := strings.TrimRight(defaultGitHubAPIBaseURL, "/")
			var tokenTTL time.Duration
			if spec.GitHubApp != nil {
				if strings.TrimSpace(spec.GitHubApp.BaseURL) != "" {
					baseURL = strings.TrimRight(spec.GitHubApp.BaseURL, "/")
				}
				tokenTTL = spec.GitHubApp.TokenTTL
			}

			provider := &appProvider{
				provider:  TypeGitHubApp,
				baseURL:   baseURL,
				tokenTTL:  tokenTTL,
				requester: httpsling.MustNew(httpsling.Client(httpclient.Timeout(defaultHTTPTimeout))),
				caps: types.ProviderCapabilities{
					SupportsRefreshTokens: true,
					SupportsClientPooling: false,
					SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
				},
			}
			provider.operations = operations.SanitizeOperationDescriptors(TypeGitHubApp, githubAppOperations(baseURL))

			return provider, nil
		},
	}
}

// appProvider implements GitHub App authentication and token minting.
type appProvider struct {
	// provider is the registered provider type.
	provider types.ProviderType
	// baseURL is the GitHub API base URL.
	baseURL string
	// tokenTTL optionally overrides installation token lifetime.
	tokenTTL time.Duration
	// requester performs HTTP requests to GitHub.
	requester *httpsling.Requester
	// caps advertises provider capabilities.
	caps types.ProviderCapabilities
	// operations enumerates supported provider operations.
	operations []types.OperationDescriptor
}

// Type returns the provider identifier.
func (p *appProvider) Type() types.ProviderType {
	return p.provider
}

// Capabilities returns the supported capabilities.
func (p *appProvider) Capabilities() types.ProviderCapabilities {
	return p.caps
}

// Operations returns the provider operation descriptors.
func (p *appProvider) Operations() []types.OperationDescriptor {
	if len(p.operations) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(p.operations))
	copy(out, p.operations)
	return out
}

// BeginAuth is not supported for GitHub App providers.
func (p *appProvider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges GitHub App credentials for an installation access token.
func (p *appProvider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	appID, installationID, privateKey, err := githubAppCredentialsFromPayload(subject.Credential)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	jwtToken, err := p.buildAppJWT(appID, privateKey)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	installToken, err := p.requestInstallationToken(ctx, installationID, jwtToken)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	payload := types.NewCredentialBuilder(p.provider).With(
		types.WithCredentialKind(types.CredentialKindOAuthToken),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: lo.Assign(map[string]any{}, subject.Credential.Data.ProviderData),
		}),
		types.WithOAuthToken(installToken),
	)

	return payload.Build()
}

// githubAppCredentialsFromPayload extracts GitHub App metadata from stored credentials.
func githubAppCredentialsFromPayload(payload types.CredentialPayload) (string, string, string, error) {
	if payload.Provider == types.ProviderUnknown {
		return "", "", "", ErrProviderNotInitialized
	}

	providerData := payload.Data.ProviderData
	appID := providerDataString(providerData, "appId")
	if appID == "" {
		return "", "", "", ErrAppIDMissing
	}

	installationID := providerDataString(providerData, "installationId")
	if installationID == "" {
		return "", "", "", ErrInstallationIDMissing
	}

	privateKey := normalizePrivateKey(providerDataString(providerData, "privateKey"))
	if privateKey == "" {
		return "", "", "", ErrPrivateKeyMissing
	}

	return appID, installationID, privateKey, nil
}

// normalizePrivateKey converts escaped newlines to PEM newlines.
func normalizePrivateKey(value string) string {
	if value == "" {
		return ""
	}

	if strings.Contains(value, "\\n") && !strings.Contains(value, "\n") {
		return strings.ReplaceAll(value, "\\n", "\n")
	}

	return value
}

// providerDataString reads a string value from provider metadata.
func providerDataString(providerData map[string]any, key string) string {
	if len(providerData) == 0 || strings.TrimSpace(key) == "" {
		return ""
	}

	value, ok := providerData[key]
	if !ok || value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		trimmed := strings.TrimSpace(fmt.Sprint(v))
		if trimmed == "<nil>" {
			return ""
		}
		return trimmed
	}
}

// buildAppJWT signs a GitHub App JWT for authentication.
func (p *appProvider) buildAppJWT(appID, privateKey string) (string, error) {
	key, err := parseRSAPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    appID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-30 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(defaultJWTExpiry)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrJWTSigningFailed, err)
	}

	return signed, nil
}

// parseRSAPrivateKey parses a PEM-encoded RSA private key.
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

// requestInstallationToken exchanges a GitHub App JWT for an installation token.
func (p *appProvider) requestInstallationToken(ctx context.Context, installationID, jwtToken string) (*oauth2.Token, error) {
	if strings.TrimSpace(installationID) == "" {
		return nil, ErrInstallationIDMissing
	}

	endpoint := fmt.Sprintf("%s/app/installations/%s/access_tokens", strings.TrimRight(p.baseURL, "/"), installationID)

	requester := p.requester
	if requester == nil {
		requester = httpsling.MustNew(httpsling.Client(httpclient.Timeout(defaultHTTPTimeout)))
	}

	var resp githubAppTokenResponse
	httpResp, err := requester.ReceiveWithContext(
		ctx,
		&resp,
		httpsling.Post(endpoint),
		httpsling.Header(httpsling.HeaderAccept, "application/vnd.github+json"),
		httpsling.BearerAuth(jwtToken),
		httpsling.JSONBody(map[string]any{}),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInstallationTokenRequestFailed, err)
	}

	if httpResp != nil {
		defer httpResp.Body.Close()
		if httpResp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("%w: %s", ErrInstallationTokenRequestFailed, httpResp.Status)
		}
	}

	if strings.TrimSpace(resp.Token) == "" {
		return nil, ErrInstallationTokenEmpty
	}

	token := &oauth2.Token{
		AccessToken: resp.Token,
		TokenType:   "Bearer",
	}

	if !resp.ExpiresAt.IsZero() {
		token.Expiry = resp.ExpiresAt
	}

	return token, nil
}

// githubAppTokenResponse models the GitHub installation token response.
type githubAppTokenResponse struct {
	// Token is the installation access token.
	Token string `json:"token"`
	// ExpiresAt is the token expiration timestamp.
	ExpiresAt time.Time `json:"expires_at"`
}
