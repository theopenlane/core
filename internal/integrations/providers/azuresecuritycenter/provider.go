package azuresecuritycenter

import (
	"context"
	"fmt"
	"maps"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
)

const (
	defaultAzureScope     = "https://management.azure.com/.default"
	azureTokenURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// Provider implements client-credential authentication for Microsoft Defender for Cloud.
type Provider struct {
	providers.BaseProvider
	tokenEndpoint func(tenantID string) string
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
	if p == nil {
		return types.CredentialPayload{}, ErrProviderNotInitialized
	}

	meta := cloneProviderData(subject.Credential.Data.ProviderData)
	if len(meta) == 0 {
		return types.CredentialPayload{}, ErrProviderMetadataRequired
	}

	credentials, err := azureSecurityCenterMetadataFromMap(meta)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	token, err := p.requestToken(ctx, credentials, subject.Scopes)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	providerData := credentials.persist(meta)
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
	if p != nil && p.tokenEndpoint != nil {
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
func defaultAzureTokenEndpoint(tenantID string) string {
	if tenantID == "" {
		return ""
	}
	return fmt.Sprintf(azureTokenURLTemplate, tenantID)
}

// cloneProviderData returns a shallow copy of provider metadata.
func cloneProviderData(data map[string]any) map[string]any {
	if len(data) == 0 {
		return nil
	}
	return maps.Clone(data)
}

type azureSecurityCenterMetadata struct {
	TenantID       string `json:"tenantId"`
	ClientID       string `json:"clientId"`
	ClientSecret   string `json:"clientSecret"`
	SubscriptionID string `json:"subscriptionId"`
	ResourceGroup  string `json:"resourceGroup"`
	WorkspaceID    string `json:"workspaceId"`
	Scope          string `json:"scope"`
}

// azureSecurityCenterMetadataFromMap normalizes and validates provider metadata.
func azureSecurityCenterMetadataFromMap(meta map[string]any) (azureSecurityCenterMetadata, error) {
	var decoded azureSecurityCenterMetadata
	if err := operations.DecodeConfig(meta, &decoded); err != nil {
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
		return []string{m.Scope}
	}
	return []string{defaultAzureScope}
}

// persist merges normalized metadata back into the provider data map.
func (m azureSecurityCenterMetadata) persist(base map[string]any) map[string]any {
	out := maps.Clone(base)
	if out == nil {
		out = map[string]any{}
	}
	out["tenantId"] = m.TenantID
	out["clientId"] = m.ClientID
	out["clientSecret"] = m.ClientSecret
	out["subscriptionId"] = m.SubscriptionID
	if m.ResourceGroup != "" {
		out["resourceGroup"] = m.ResourceGroup
	}
	if m.WorkspaceID != "" {
		out["workspaceId"] = m.WorkspaceID
	}
	if m.Scope != "" {
		out["scope"] = m.Scope
	}
	return out
}
