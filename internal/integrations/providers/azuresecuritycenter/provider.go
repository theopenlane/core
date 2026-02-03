package azuresecuritycenter

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

const (
	defaultAzureScope     = "https://management.azure.com/.default"
	azureTokenURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// Provider implements client-credential authentication for Microsoft Defender for Cloud.
type Provider struct {
	provider      types.ProviderType
	operations    []types.OperationDescriptor
	caps          types.ProviderCapabilities
	tokenEndpoint func(tenantID string) string
	clients       []types.ClientDescriptor
}

// newProvider constructs the Azure Security Center provider from a spec.
func newProvider(spec config.ProviderSpec) *Provider {
	return &Provider{
		provider:   TypeAzureSecurityCenter,
		operations: helpers.SanitizeOperationDescriptors(TypeAzureSecurityCenter, azureSecurityOperations()),
		caps: types.ProviderCapabilities{
			SupportsRefreshTokens: true,
			SupportsClientPooling: true,
			SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
		},
		tokenEndpoint: defaultAzureTokenEndpoint,
		clients:       helpers.SanitizeClientDescriptors(TypeAzureSecurityCenter, azureSecurityCenterClientDescriptors()),
	}
}

// Type returns the provider identifier.
func (p *Provider) Type() types.ProviderType {
	if p == nil {
		return types.ProviderUnknown
	}
	return p.provider
}

// Capabilities returns the provider capabilities.
func (p *Provider) Capabilities() types.ProviderCapabilities {
	if p == nil {
		return types.ProviderCapabilities{}
	}
	return p.caps
}

// Operations returns provider-published operations.
func (p *Provider) Operations() []types.OperationDescriptor {
	if p == nil || len(p.operations) == 0 {
		return nil
	}
	out := make([]types.OperationDescriptor, len(p.operations))
	copy(out, p.operations)
	return out
}

// ClientDescriptors returns provider-published client descriptors when configured.
func (p *Provider) ClientDescriptors() []types.ClientDescriptor {
	if p == nil || len(p.clients) == 0 {
		return nil
	}

	out := make([]types.ClientDescriptor, len(p.clients))
	copy(out, p.clients)
	return out
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
	builder := types.NewCredentialBuilder(p.provider).With(
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
	tenantID = strings.TrimSpace(tenantID)
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
	TenantID       string
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	ResourceGroup  string
	WorkspaceID    string
	Scope          string
}

// azureSecurityCenterMetadataFromMap normalizes and validates provider metadata.
func azureSecurityCenterMetadataFromMap(meta map[string]any) (azureSecurityCenterMetadata, error) {
	tenantID, err := helpers.RequiredString(meta, "tenantId", ErrTenantIDMissing)
	if err != nil {
		return azureSecurityCenterMetadata{}, err
	}
	clientID, err := helpers.RequiredString(meta, "clientId", ErrClientIDMissing)
	if err != nil {
		return azureSecurityCenterMetadata{}, err
	}
	clientSecret, err := helpers.RequiredString(meta, "clientSecret", ErrClientSecretMissing)
	if err != nil {
		return azureSecurityCenterMetadata{}, err
	}
	subscriptionID, err := helpers.RequiredString(meta, "subscriptionId", ErrSubscriptionIDMissing)
	if err != nil {
		return azureSecurityCenterMetadata{}, err
	}

	return azureSecurityCenterMetadata{
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		SubscriptionID: subscriptionID,
		ResourceGroup:  helpers.StringValue(meta, "resourceGroup"),
		WorkspaceID:    helpers.StringValue(meta, "workspaceId"),
		Scope:          helpers.StringValue(meta, "scope"),
	}, nil
}

// scopes returns the scopes to request for the client credentials flow.
func (m azureSecurityCenterMetadata) scopes(overrides []string) []string {
	if len(overrides) > 0 {
		return append([]string(nil), overrides...)
	}
	if strings.TrimSpace(m.Scope) != "" {
		return []string{strings.TrimSpace(m.Scope)}
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
	if strings.TrimSpace(m.Scope) != "" {
		out["scope"] = strings.TrimSpace(m.Scope)
	}
	return out
}
