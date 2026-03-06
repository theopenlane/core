package github

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gh "github.com/google/go-github/v83/github"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// TypeGitHubApp identifies the GitHub App provider
	TypeGitHubApp = types.ProviderType("githubapp")
)

var _ types.ClientProvider = (*appProvider)(nil)

// Default GitHub App API configuration values.
const (
	defaultJWTExpiry = 9 * time.Minute
)

// AppBuilder returns the GitHub App provider builder.
func AppBuilder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGitHubApp,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.AuthType != "" && spec.AuthType != types.AuthKindGitHubApp {
				return nil, fmt.Errorf("%w (provider %s expects %s, found %s)", ErrAuthTypeMismatch, TypeGitHubApp, types.AuthKindGitHubApp, spec.AuthType)
			}

			baseURL := strings.TrimRight(githubAPIBaseURL, "/")
			appID := ""
			privateKey := ""
			var tokenTTL time.Duration
			if spec.GitHubApp != nil {
				if strings.TrimSpace(spec.GitHubApp.BaseURL) != "" {
					baseURL = strings.TrimRight(spec.GitHubApp.BaseURL, "/")
				}
				tokenTTL = spec.GitHubApp.TokenTTL
				appID = spec.GitHubApp.AppID
				privateKey = normalizePrivateKey(spec.GitHubApp.PrivateKey)
			}

			clients := providerkit.SanitizeClientDescriptors(TypeGitHubApp, githubClientDescriptorsWithBaseURL(TypeGitHubApp, baseURL))

			provider := &appProvider{
				provider:   TypeGitHubApp,
				baseURL:    baseURL,
				appID:      appID,
				privateKey: privateKey,
				tokenTTL:   tokenTTL,
				clients:    clients,
				caps: types.ProviderCapabilities{
					SupportsRefreshTokens:  true,
					SupportsClientPooling:  len(clients) > 0,
					SupportsMetadataForm:   len(spec.CredentialsSchema) > 0,
					EnvironmentCredentials: true,
				},
			}
			provider.operations = providerkit.SanitizeOperationDescriptors(TypeGitHubApp, githubAppOperations(baseURL))

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
	// appID is the runtime GitHub App identifier used for JWT signing.
	appID string
	// privateKey is the runtime GitHub App private key used for JWT signing.
	privateKey string
	// tokenTTL optionally overrides installation token lifetime.
	tokenTTL time.Duration
	// caps advertises provider capabilities.
	caps types.ProviderCapabilities
	// clients enumerates supported pooled clients.
	clients []types.ClientDescriptor
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

// ClientDescriptors returns provider-published client descriptors.
func (p *appProvider) ClientDescriptors() []types.ClientDescriptor {
	if len(p.clients) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, len(p.clients))
	copy(out, p.clients)

	return out
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

// DefaultMappings returns the built-in ingest mapping registrations for GitHub App providers.
func (p *appProvider) DefaultMappings() []types.MappingRegistration {
	return githubDefaultMappings()
}

// Mint exchanges GitHub App credentials for an installation access token.
func (p *appProvider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	appID, installationID, privateKey, err := p.resolveMintInputs(subject.Credential)
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

	providerData := maps.Clone(subject.Credential.Data.ProviderData)
	if providerData == nil {
		providerData = map[string]any{}
	}

	providerData["appId"] = appID
	providerData["installationId"] = installationID
	delete(providerData, "privateKey")

	payload := types.NewCredentialBuilder(p.provider).With(
		types.WithCredentialKind(types.CredentialKindOAuthToken),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: providerData,
		}),
		types.WithOAuthToken(installToken),
	)

	minted, err := payload.Build()
	if err != nil {
		return types.CredentialPayload{}, err
	}

	if subject.Credential.ProviderState != nil {
		minted.ProviderState = subject.Credential.ProviderState
	}

	return minted, nil
}

// githubAppCredentialsFromPayload extracts GitHub App metadata from stored credentials.
func githubAppCredentialsFromPayload(payload types.CredentialPayload) (string, string, string, error) {
	if payload.Provider == types.ProviderUnknown {
		return "", "", "", ErrProviderNotInitialized
	}

	decoded, err := githubAppProviderDataFromPayload(payload)
	if err != nil {
		return "", "", "", err
	}

	if decoded.AppID == "" {
		return "", "", "", ErrAppIDMissing
	}

	installationID, err := resolveInstallationID(decoded, payload)
	if err != nil {
		return "", "", "", err
	}

	privateKey := normalizePrivateKey(decoded.PrivateKey.String())
	if privateKey == "" {
		return "", "", "", ErrPrivateKeyMissing
	}

	return decoded.AppID.String(), installationID, privateKey, nil
}

func (p *appProvider) resolveMintInputs(payload types.CredentialPayload) (string, string, string, error) {
	installationID, err := githubAppInstallationIDFromCredential(payload)
	if err != nil {
		return "", "", "", err
	}

	appID := p.appID
	if appID == "" {
		return "", "", "", ErrAppIDMissing
	}

	privateKey := normalizePrivateKey(p.privateKey)
	if privateKey == "" {
		return "", "", "", ErrPrivateKeyMissing
	}

	return appID, installationID, privateKey, nil
}

func githubAppProviderDataFromPayload(payload types.CredentialPayload) (githubAppProviderData, error) {
	var decoded githubAppProviderData
	if err := auth.DecodeProviderData(payload.Data.ProviderData, &decoded); err != nil {
		return githubAppProviderData{}, err
	}

	return decoded, nil
}

// resolveInstallationID returns the installation ID from the already-decoded provider data,
// falling back to the provider state when the credential payload does not carry it directly.
func resolveInstallationID(decoded githubAppProviderData, payload types.CredentialPayload) (string, error) {
	if decoded.InstallationID != "" {
		return decoded.InstallationID.String(), nil
	}

	if payload.ProviderState == nil {
		return "", ErrInstallationIDMissing
	}

	stateProviderData, err := payload.ProviderState.ProviderDataMap(string(TypeGitHubApp))
	if err != nil {
		return "", err
	}

	var stateDecoded githubAppProviderData
	if err := auth.DecodeProviderData(stateProviderData, &stateDecoded); err != nil {
		return "", err
	}

	if stateDecoded.InstallationID == "" {
		return "", ErrInstallationIDMissing
	}

	return stateDecoded.InstallationID.String(), nil
}

func githubAppInstallationIDFromCredential(payload types.CredentialPayload) (string, error) {
	decoded, err := githubAppProviderDataFromPayload(payload)
	if err != nil {
		return "", err
	}

	return resolveInstallationID(decoded, payload)
}

// normalizePrivateKey converts escaped newlines to PEM newlines.
func normalizePrivateKey(value string) string {
	if strings.Contains(value, "\\n") && !strings.Contains(value, "\n") {
		return strings.ReplaceAll(value, "\\n", "\n")
	}

	return value
}

type githubAppProviderData struct {
	AppID          types.TrimmedString `json:"appId"`
	InstallationID types.TrimmedString `json:"installationId"`
	PrivateKey     types.TrimmedString `json:"privateKey"`
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
	if installationID == "" {
		return nil, ErrInstallationIDMissing
	}

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInstallationTokenRequestFailed, err)
	}

	client, err := newGitHubAPIClient(jwtToken, p.baseURL)
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
