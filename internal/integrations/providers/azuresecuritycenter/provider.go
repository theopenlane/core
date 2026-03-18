package azuresecuritycenter

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
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
			operations.SanitizeOperationDescriptors(TypeAzureSecurityCenter, azureSecurityOperations()),
			operations.SanitizeClientDescriptors(TypeAzureSecurityCenter, azureSecurityCenterClientDescriptors()),
		),
		tokenEndpoint: defaultAzureTokenEndpoint,
	}
}

// BeginAuth is not supported for Azure Security Center client credentials.
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges stored client credentials for an Azure access token.
func (p *Provider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	if len(subject.Credential.Data.ProviderData) == 0 {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	credentials, err := azureSecurityCenterMetadataFromMap(subject.Credential.Data.ProviderData)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	token, err := p.requestToken(ctx, credentials, subject.Scopes)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	providerData, err := auth.PersistMetadata(subject.Credential.Data.ProviderData, credentials)
	if err != nil {
		return types.CredentialPayload{}, err
	}
	builder := types.NewCredentialBuilder(p.Type()).With(
		types.WithCredentialKind(types.CredentialKindOAuthToken),
		types.WithOAuthToken(token),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: providerData,
		}),
	)

	return builder.Build()
}

// requestToken obtains an Azure access token using the client credentials flow.
func (p *Provider) requestToken(ctx context.Context, meta azureSecurityCenterMetadata, scopes []string) (*oauth2.Token, error) {
	tokenURL := defaultAzureTokenEndpoint(meta.TenantID)
	if p.tokenEndpoint != nil {
		tokenURL = p.tokenEndpoint(meta.TenantID)
	}

	scopeList := meta.scopes(scopes)
	cfg := clientcredentials.Config{
		ClientID:     meta.ClientID.String(),
		ClientSecret: meta.ClientSecret.String(),
		TokenURL:     tokenURL,
		Scopes:       scopeList,
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	token, err := cfg.Token(ctx)
	if err != nil {
		return nil, ErrTokenExchangeFailed
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
	ClientID types.TrimmedString `json:"clientId,omitempty"`
	// ClientSecret holds the client credential secret
	ClientSecret types.TrimmedString `json:"clientSecret,omitempty"`
	// SubscriptionID identifies the Azure subscription
	SubscriptionID types.TrimmedString `json:"subscriptionId,omitempty"`
	// ResourceGroup scopes access to a resource group
	ResourceGroup types.TrimmedString `json:"resourceGroup,omitempty"`
	// WorkspaceID identifies the Defender workspace
	WorkspaceID types.TrimmedString `json:"workspaceId,omitempty"`
	// Scope overrides the default OAuth scope
	Scope types.TrimmedString `json:"scope,omitempty"`
}

// azureSecurityCenterMetadataFromMap normalizes and validates provider metadata.
func azureSecurityCenterMetadataFromMap(meta map[string]any) (azureSecurityCenterMetadata, error) {
	var decoded azureSecurityCenterMetadata
	if err := auth.DecodeProviderData(meta, &decoded); err != nil {
		return azureSecurityCenterMetadata{}, err
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
