package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
)

// TypeGitHubApp identifies the GitHub App provider.
const TypeGitHubApp = types.ProviderType("github_app")

// AppBuilder returns the GitHub App provider builder.
func AppBuilder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGitHubApp,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.AuthType != "" && spec.AuthType != types.AuthKindGitHubApp {
				return nil, ErrAuthTypeMismatch
			}

			return &AppProvider{
				BaseProvider: providers.NewBaseProvider(
					TypeGitHubApp,
					types.ProviderCapabilities{
						SupportsRefreshTokens: true,
						SupportsClientPooling: true,
						SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
					},
					operations.SanitizeOperationDescriptors(TypeGitHubApp, githubOperations()),
					operations.SanitizeClientDescriptors(TypeGitHubApp, githubClientDescriptors(TypeGitHubApp)),
				),
			}, nil
		},
	}
}

// AppProvider implements GitHub App authentication via installation tokens.
type AppProvider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
}

// BeginAuth is not supported for GitHub App providers.
func (p *AppProvider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges the stored GitHub App credentials for a short-lived installation token.
func (p *AppProvider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	state := subject.Credential.ProviderState
	if state == nil || state.GitHub == nil {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	appID := state.GitHub.AppID
	if appID == "" {
		return types.CredentialPayload{}, ErrAppIDMissing
	}

	installationID := state.GitHub.InstallationID
	if installationID == "" {
		return types.CredentialPayload{}, ErrInstallationIDMissing
	}

	meta := subject.Credential.Data.ProviderData
	if len(meta) == 0 {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	var decoded struct {
		// AppID identifies the GitHub App ID
		AppID string `json:"appId"`
		// InstallationID identifies the GitHub App installation
		InstallationID string `json:"installationId"`
		// PrivateKey holds the GitHub App private key
		PrivateKey string `json:"privateKey"`
	}

	if err := operations.DecodeConfig(meta, &decoded); err != nil {
		return types.CredentialPayload{}, err
	}
	privateKey := decoded.PrivateKey
	if privateKey == "" {
		return types.CredentialPayload{}, ErrPrivateKeyMissing
	}

	jwtToken, err := buildGitHubAppJWT(appID, privateKey)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	resp, err := requestGitHubInstallationToken(ctx, jwtToken, installationID)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	oauthToken := &oauth2.Token{
		AccessToken: resp.Token,
		TokenType:   "Bearer",
	}
	if !resp.ExpiresAt.IsZero() {
		oauthToken.Expiry = resp.ExpiresAt
	}

	cloned := maps.Clone(meta)
	builder := types.NewCredentialBuilder(p.Type()).With(
		types.WithOAuthToken(oauthToken),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: cloned,
		}),
	)

	return builder.Build()
}

// buildGitHubAppJWT signs a JWT for GitHub App authentication
func buildGitHubAppJWT(appID string, privateKey string) (string, error) {
	privateKey = normalizeGitHubPrivateKey(privateKey)
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", ErrAppPrivateKeyParse
	}

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    appID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", ErrAppJWTSign
	}

	return signed, nil
}

// normalizeGitHubPrivateKey normalizes and decodes private key input
func normalizeGitHubPrivateKey(value string) string {
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\\n", "\n")
	if strings.HasPrefix(value, "-----BEGIN") {
		return value
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return string(decoded)
	}

	return value
}

// githubInstallationTokenResponse captures the response from GitHub's installation token endpoint
type githubInstallationTokenResponse struct {
	// Token is the installation access token
	Token string `json:"token"`
	// ExpiresAt is the token expiration time
	ExpiresAt time.Time `json:"expires_at"`
}

// requestGitHubInstallationToken exchanges a JWT for an installation token
func requestGitHubInstallationToken(ctx context.Context, jwtToken string, installationID string) (githubInstallationTokenResponse, error) {
	if jwtToken == "" {
		return githubInstallationTokenResponse{}, ErrAppJWTMissing
	}
	if installationID == "" {
		return githubInstallationTokenResponse{}, ErrInstallationIDMissing
	}

	path := fmt.Sprintf("app/installations/%s/access_tokens", installationID)
	endpoint := githubAPIBaseURL + path

	headers := map[string]string{
		"Accept":               "application/vnd.github+json",
		"X-GitHub-Api-Version": githubAPIVersion,
	}

	var resp githubInstallationTokenResponse
	if err := auth.HTTPPostJSON(ctx, nil, endpoint, jwtToken, headers, map[string]any{}, &resp); err != nil {
		if errors.Is(err, auth.ErrHTTPRequestFailed) {
			return githubInstallationTokenResponse{}, ErrAPIRequest
		}
		return githubInstallationTokenResponse{}, err
	}

	return resp, nil
}
