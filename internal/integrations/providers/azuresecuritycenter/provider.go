package azuresecuritycenter

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	defaultAzureScope     = "https://management.azure.com/.default"
	azureTokenURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token" //nolint:gosec // G101: URL template, not credentials
)

// Provider implements client-credential authentication for Microsoft Defender for Cloud.
type Provider struct {
	// BaseProvider provides shared provider metadata
	providers.BaseProvider
	tokenEndpoint func(tenantID types.TrimmedString) string
}

// newProvider constructs the Azure Security Center provider from a spec.
func newProvider(spec config.ProviderSpec) *Provider {
	return &Provider{
		BaseProvider: providers.NewBaseProvider(
			TypeAzureSecurityCenter,
			types.ProviderCapabilities{
				SupportsRefreshTokens: true,
				SupportsClientPooling: true,
				SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
			},
			providerkit.SanitizeOperationDescriptors(TypeAzureSecurityCenter, azureSecurityOperations()),
			providerkit.SanitizeClientDescriptors(TypeAzureSecurityCenter, azureSecurityCenterClientDescriptors()),
		),
		tokenEndpoint: defaultAzureTokenEndpoint,
	}
}

// BeginAuth is not supported for Azure Security Center client credentials.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges stored client credentials for an Azure access token.
func (p *Provider) Mint(ctx context.Context, subject types.CredentialMintRequest) (models.CredentialSet, error) {
	credentials, err := azureSecurityCenterMetadataFromPayload(subject.Credential)
	if err != nil {
		return models.CredentialSet{}, err
	}

	token, err := p.requestToken(ctx, credentials, subject.Scopes)
	if err != nil {
		return models.CredentialSet{}, err
	}

	providerData, err := jsonx.ToRawMessage(credentials.providerData())
	if err != nil {
		return models.CredentialSet{}, err
	}

	credential := models.CredentialSet{
		ClientID:          credentials.ClientID,
		ClientSecret:      credentials.ClientSecret,
		ProviderData:      providerData,
		OAuthAccessToken:  token.AccessToken,
		OAuthRefreshToken: token.RefreshToken,
		OAuthTokenType:    token.TokenType,
	}
	if !token.Expiry.IsZero() {
		exp := token.Expiry.UTC()
		credential.OAuthExpiry = &exp
	}

	return credential, nil
}

// requestToken obtains an Azure access token using the client credentials flow.
func (p *Provider) requestToken(ctx context.Context, meta azureSecurityCenterMetadata, scopes []string) (*oauth2.Token, error) {
	tokenURL := defaultAzureTokenEndpoint(meta.TenantID)
	if p.tokenEndpoint != nil {
		tokenURL = p.tokenEndpoint(meta.TenantID)
	}

	scopeList := meta.scopes(scopes)
	cfg := clientcredentials.Config{
		ClientID:     meta.ClientID,
		ClientSecret: meta.ClientSecret,
		TokenURL:     tokenURL,
		Scopes:       scopeList,
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	token, err := cfg.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenExchangeFailed, err)
	}

	return token, nil
}

// defaultAzureTokenEndpoint builds the Azure token endpoint for the tenant.
func defaultAzureTokenEndpoint(tenantID types.TrimmedString) string {
	if tenantID == "" {
		return ""
	}
	return fmt.Sprintf(azureTokenURLTemplate, tenantID)
}

type azureSecurityCenterMetadata struct {
	// TenantID identifies the Azure tenant
	TenantID types.TrimmedString `json:"tenantId,omitempty"`
	// ClientID identifies the Azure application
	ClientID string `json:"clientId,omitempty"`
	// ClientSecret holds the client credential secret
	ClientSecret string `json:"clientSecret,omitempty"`
	// SubscriptionID identifies the Azure subscription
	SubscriptionID types.TrimmedString `json:"subscriptionId,omitempty"`
	// ResourceGroup scopes access to a resource group
	ResourceGroup types.TrimmedString `json:"resourceGroup,omitempty"`
	// WorkspaceID identifies the Defender workspace
	WorkspaceID types.TrimmedString `json:"workspaceId,omitempty"`
	// Scope overrides the default OAuth scope
	Scope types.TrimmedString `json:"scope,omitempty"`
}

type azureSecurityCenterProviderData struct {
	TenantID       string `json:"tenantId,omitempty"`
	SubscriptionID string `json:"subscriptionId,omitempty"`
	ResourceGroup  string `json:"resourceGroup,omitempty"`
	WorkspaceID    string `json:"workspaceId,omitempty"`
	Scope          string `json:"scope,omitempty"`
}

func (m azureSecurityCenterMetadata) providerData() azureSecurityCenterProviderData {
	return azureSecurityCenterProviderData{
		TenantID:       m.TenantID.String(),
		SubscriptionID: m.SubscriptionID.String(),
		ResourceGroup:  m.ResourceGroup.String(),
		WorkspaceID:    m.WorkspaceID.String(),
		Scope:          m.Scope.String(),
	}
}

// azureSecurityCenterMetadataFromPayload normalizes and validates provider metadata.
func azureSecurityCenterMetadataFromPayload(payload models.CredentialSet) (azureSecurityCenterMetadata, error) {
	if len(payload.ProviderData) == 0 && payload.ClientID == "" && payload.ClientSecret == "" {
		return azureSecurityCenterMetadata{}, ErrProviderMetadataRequired
	}

	var decoded azureSecurityCenterMetadata
	if err := jsonx.UnmarshalIfPresent(payload.ProviderData, &decoded); err != nil {
		return azureSecurityCenterMetadata{}, err
	}
	if decoded.ClientID == "" {
		decoded.ClientID = payload.ClientID
	}
	if decoded.ClientSecret == "" {
		decoded.ClientSecret = payload.ClientSecret
	}

	switch {
	case decoded.TenantID == "":
		return azureSecurityCenterMetadata{}, ErrTenantIDMissing
	case decoded.ClientID == "":
		return azureSecurityCenterMetadata{}, ErrClientIDMissing
	case decoded.ClientSecret == "":
		return azureSecurityCenterMetadata{}, ErrClientSecretMissing
	case decoded.SubscriptionID == "":
		return azureSecurityCenterMetadata{}, ErrSubscriptionIDMissing
	}

	return decoded, nil
}

// scopes returns the scopes to request for the client credentials flow.
func (m azureSecurityCenterMetadata) scopes(overrides []string) []string {
	if len(overrides) > 0 {
		return append([]string(nil), overrides...)
	}
	if m.Scope != "" {
		return []string{m.Scope.String()}
	}

	return []string{defaultAzureScope}
}
