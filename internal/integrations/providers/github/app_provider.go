package github

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
	"github.com/samber/lo"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// TypeGitHubApp identifies the GitHub App provider
	TypeGitHubApp = types.ProviderType("githubapp")

	// defaultJWTExpiry is the lifetime of a GitHub App JWT
	defaultJWTExpiry = 9 * time.Minute

	// jwtIssuedAtBackdateSeconds is subtracted from now when setting IssuedAt to allow for clock skew
	jwtIssuedAtBackdateSeconds = 30 * time.Second
)

var _ types.ClientProvider = (*appProvider)(nil)

// githubAppProviderData captures provider-specific data stored in a credential set.
type githubAppProviderData struct {
	// AppID is the GitHub App identifier
	AppID string `json:"appId"`
	// InstallationID is the GitHub App installation identifier
	InstallationID string `json:"installationId"`
	// PrivateKey is not stored after minting (cleared before persisting)
	PrivateKey string `json:"privateKey"`
}

// AppBuilder returns the GitHub App provider builder with the supplied operator config applied.
func AppBuilder(cfg AppConfig) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGitHubApp,
		SpecFunc:     githubAppSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if err := providerkit.ValidateAuthType(s, types.AuthKindGitHubApp, ErrAuthTypeMismatch); err != nil {
				return nil, err
			}

			baseURL := strings.TrimRight(githubAPIBaseURL, "/")
			if strings.TrimSpace(cfg.BaseURL) != "" {
				baseURL = strings.TrimRight(cfg.BaseURL, "/")
			}

			privateKey := normalizePrivateKey(cfg.PrivateKey)

			clients := providerkit.SanitizeClientDescriptors(TypeGitHubApp, githubClientDescriptorsWithBaseURL(TypeGitHubApp, baseURL))

			p := &appProvider{
				provider:   TypeGitHubApp,
				baseURL:    baseURL,
				appID:      cfg.AppID,
				privateKey: privateKey,
				tokenTTL:   cfg.TokenTTL,
				clients:    clients,
				caps: types.ProviderCapabilities{
					SupportsRefreshTokens:  true,
					SupportsClientPooling:  len(clients) > 0,
					SupportsMetadataForm:   len(s.CredentialsSchema) > 0,
					EnvironmentCredentials: true,
				},
			}
			p.operations = providerkit.SanitizeOperationDescriptors(TypeGitHubApp, githubAppOperations(baseURL))

			return p, nil
		},
	}
}

// githubAppSpec returns the static provider specification for the GitHub App provider.
func githubAppSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "githubapp",
		DisplayName:      "GitHub App",
		Category:         "code",
		AuthType:         types.AuthKindGitHubApp,
		AuthStartPath:    "/v1/integrations/github/app/install",
		AuthCallbackPath: "/v1/integrations/github/app/callback",
		Active:           lo.ToPtr(true),
		Visible:          lo.ToPtr(true),
		LogoURL:          "",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/github_app/overview",
		Description:      "Install the Openlane GitHub App to collect repository metadata and security alerts (Dependabot, code scanning, and secret scanning) for exposure management.",
	}
}

// appProvider implements GitHub App authentication and token minting.
type appProvider struct {
	// provider is the registered provider type
	provider types.ProviderType
	// baseURL is the GitHub API base URL
	baseURL string
	// appID is the runtime GitHub App identifier used for JWT signing
	appID string
	// privateKey is the runtime GitHub App private key used for JWT signing
	privateKey string
	// tokenTTL optionally overrides installation token lifetime
	tokenTTL time.Duration
	// caps advertises provider capabilities
	caps types.ProviderCapabilities
	// clients enumerates supported pooled clients
	clients []types.ClientDescriptor
	// operations enumerates supported provider operations
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
func (p *appProvider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// DefaultMappings returns the built-in ingest mapping registrations for GitHub App providers.
func (p *appProvider) DefaultMappings() []types.MappingRegistration {
	return githubDefaultMappings()
}

// Mint exchanges GitHub App credentials for an installation access token.
func (p *appProvider) Mint(ctx context.Context, subject types.CredentialMintRequest) (types.CredentialSet, error) {
	appID, installationID, privateKey, err := p.resolveMintInputs(subject.Credential, subject.ProviderState)
	if err != nil {
		return types.CredentialSet{}, err
	}

	jwtToken, err := p.buildAppJWT(appID, privateKey)
	if err != nil {
		return types.CredentialSet{}, err
	}

	installToken, err := p.requestInstallationToken(ctx, installationID, jwtToken)
	if err != nil {
		return types.CredentialSet{}, err
	}

	metadata, err := githubAppProviderDataFromCredential(subject.Credential)
	if err != nil {
		return types.CredentialSet{}, err
	}

	metadata.AppID = appID
	metadata.InstallationID = installationID
	metadata.PrivateKey = ""

	providerData, err := jsonx.ToRawMessage(metadata)
	if err != nil {
		return types.CredentialSet{}, err
	}

	credential := types.CredentialSet{
		ProviderData:      providerData,
		OAuthAccessToken:  installToken.AccessToken,
		OAuthRefreshToken: installToken.RefreshToken,
		OAuthTokenType:    installToken.TokenType,
	}

	if !installToken.Expiry.IsZero() {
		exp := installToken.Expiry.UTC()
		credential.OAuthExpiry = &exp
	}

	return credential, nil
}

func (p *appProvider) resolveMintInputs(credential types.CredentialSet, providerState json.RawMessage) (string, string, string, error) {
	installationID, err := githubAppInstallationIDFromCredential(credential, providerState)
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

func githubAppProviderDataFromCredential(credential types.CredentialSet) (githubAppProviderData, error) {
	var decoded githubAppProviderData
	if err := jsonx.UnmarshalIfPresent(credential.ProviderData, &decoded); err != nil {
		return githubAppProviderData{}, err
	}

	return decoded, nil
}

// resolveInstallationID returns the installation ID from the already-decoded provider data,
// falling back to the provider state when the credential does not carry it directly.
func resolveInstallationID(decoded githubAppProviderData, providerState json.RawMessage) (string, error) {
	if decoded.InstallationID != "" {
		return decoded.InstallationID, nil
	}

	if len(providerState) == 0 {
		return "", ErrInstallationIDMissing
	}

	// Decode provider state — try {"providers": {"githubapp": {...}}} format
	type providerStateEnvelope struct {
		Providers map[string]json.RawMessage `json:"providers"`
	}

	var envelope providerStateEnvelope

	var stateRaw json.RawMessage
	if err := jsonx.UnmarshalIfPresent(providerState, &envelope); err == nil {
		stateRaw = envelope.Providers[string(TypeGitHubApp)]
	}

	if len(stateRaw) == 0 {
		return "", ErrInstallationIDMissing
	}

	var stateDecoded githubAppProviderData
	if err := jsonx.UnmarshalIfPresent(stateRaw, &stateDecoded); err != nil {
		return "", err
	}

	if stateDecoded.InstallationID == "" {
		return "", ErrInstallationIDMissing
	}

	return stateDecoded.InstallationID, nil
}

func githubAppInstallationIDFromCredential(credential types.CredentialSet, providerState json.RawMessage) (string, error) {
	decoded, err := githubAppProviderDataFromCredential(credential)
	if err != nil {
		return "", err
	}

	return resolveInstallationID(decoded, providerState)
}

// normalizePrivateKey converts escaped newlines to PEM newlines.
func normalizePrivateKey(value string) string {
	if strings.Contains(value, "\\n") && !strings.Contains(value, "\n") {
		return strings.ReplaceAll(value, "\\n", "\n")
	}

	return value
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

	client, err := newGitHubAPIClient(ctx, jwtToken, p.baseURL)
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
