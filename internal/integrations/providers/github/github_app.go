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

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/integrations/providers"
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
				provider:   TypeGitHubApp,
				operations: helpers.SanitizeOperationDescriptors(TypeGitHubApp, githubOperations()),
				caps: types.ProviderCapabilities{
					SupportsRefreshTokens: true,
					SupportsClientPooling: true,
					SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
				},
				clients: helpers.SanitizeClientDescriptors(TypeGitHubApp, githubClientDescriptors(TypeGitHubApp)),
			}, nil
		},
	}
}

// AppProvider implements GitHub App authentication via installation tokens.
type AppProvider struct {
	provider   types.ProviderType
	operations []types.OperationDescriptor
	caps       types.ProviderCapabilities
	clients    []types.ClientDescriptor
}

// Type returns the provider identifier.
func (p *AppProvider) Type() types.ProviderType {
	if p == nil {
		return types.ProviderUnknown
	}
	return p.provider
}

// Capabilities returns optional capability flags.
func (p *AppProvider) Capabilities() types.ProviderCapabilities {
	if p == nil {
		return types.ProviderCapabilities{}
	}
	return p.caps
}

// Operations returns provider-published operations.
func (p *AppProvider) Operations() []types.OperationDescriptor {
	if p == nil || len(p.operations) == 0 {
		return nil
	}

	out := make([]types.OperationDescriptor, len(p.operations))
	copy(out, p.operations)
	return out
}

// ClientDescriptors returns provider-published client descriptors when configured.
func (p *AppProvider) ClientDescriptors() []types.ClientDescriptor {
	if p == nil || len(p.clients) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, len(p.clients))
	copy(out, p.clients)
	return out
}

// BeginAuth is not supported for GitHub App providers.
func (p *AppProvider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges the stored GitHub App credentials for a short-lived installation token.
func (p *AppProvider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	if p == nil {
		return types.CredentialPayload{}, ErrProviderNotInitialized
	}

	meta := cloneGitHubProviderData(subject.Credential.Data.ProviderData)
	if len(meta) == 0 {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	appID := helpers.FirstStringValue(meta, "appId", "app_id")
	if appID == "" {
		return types.CredentialPayload{}, ErrAppIDMissing
	}

	installationID := helpers.FirstStringValue(meta, "installationId", "installation_id")
	if installationID == "" {
		return types.CredentialPayload{}, ErrInstallationIDMissing
	}

	privateKey := helpers.FirstStringValue(meta, "privateKey", "private_key")
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
	builder := types.NewCredentialBuilder(p.provider).With(
		types.WithOAuthToken(oauthToken),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: cloned,
		}),
	)

	return builder.Build()
}

func cloneGitHubProviderData(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}

	return maps.Clone(data)
}

func buildGitHubAppJWT(appID string, privateKey string) (string, error) {
	privateKey = normalizeGitHubPrivateKey(privateKey)
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAppPrivateKeyParse, err)
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
		return "", fmt.Errorf("%w: %w", ErrAppJWTSign, err)
	}

	return signed, nil
}

func normalizeGitHubPrivateKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\\n", "\n")
	if strings.HasPrefix(value, "-----BEGIN") {
		return value
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return strings.TrimSpace(string(decoded))
	}
	return value
}

type githubInstallationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func requestGitHubInstallationToken(ctx context.Context, jwtToken string, installationID string) (githubInstallationTokenResponse, error) {
	if strings.TrimSpace(jwtToken) == "" {
		return githubInstallationTokenResponse{}, errors.New("github: app jwt missing")
	}
	installationID = strings.TrimSpace(installationID)
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
	if err := helpers.HTTPPostJSON(ctx, nil, endpoint, jwtToken, headers, map[string]any{}, &resp); err != nil {
		if errors.Is(err, helpers.ErrHTTPRequestFailed) {
			return githubInstallationTokenResponse{}, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}
		return githubInstallationTokenResponse{}, err
	}

	return resp, nil
}
