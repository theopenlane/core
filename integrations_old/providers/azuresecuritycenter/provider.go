package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// azureSecurityCenterCredentialsSchema is the JSON Schema for Azure Security Center credentials.
var azureSecurityCenterCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["tenantId","clientId","clientSecret","subscriptionId"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this Defender for Cloud deployment."},"tenantId":{"type":"string","title":"Tenant ID","description":"Azure AD tenant ID that owns the Defender for Cloud resources."},"clientId":{"type":"string","title":"Client ID","description":"Application (client) ID of the service principal used for API calls."},"clientSecret":{"type":"string","title":"Client Secret","description":"Client secret associated with the service principal."},"subscriptionId":{"type":"string","title":"Subscription ID","description":"Azure subscription that hosts Defender for Cloud."},"resourceGroup":{"type":"string","title":"Resource Group","description":"Optional resource group for more granular scoping."},"workspaceId":{"type":"string","title":"Log Analytics Workspace ID","description":"Optional workspace ID when querying Defender alerts via Log Analytics."}}}`)

// TypeAzureSecurityCenter identifies the Azure Security Center provider
const TypeAzureSecurityCenter = types.ProviderType("azuresecuritycenter")

const (
	// ClientAzureSecurityCenterAPI identifies the Azure management API client
	ClientAzureSecurityCenterAPI types.ClientName = "api"

	azureSubscriptionScopePrefix = "subscriptions/"
	defaultAzureScope            = "https://management.azure.com/.default"
	azureTokenURLTemplate        = "https://login.microsoftonline.com/%s/oauth2/v2.0/token" //nolint:gosec // G101: URL template, not credentials
)

const (
	azureSecurityHealth  types.OperationName = types.OperationHealthDefault
	azureSecurityPricing types.OperationName = "security.pricing_overview"
)

// azureSecurityCenterMetadata holds provider-specific credential fields decoded from ProviderData
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

// scopes returns the scopes to request for the client credentials flow
func (m azureSecurityCenterMetadata) scopes(overrides []string) []string {
	if len(overrides) > 0 {
		return append([]string(nil), overrides...)
	}

	if m.Scope != "" {
		return []string{m.Scope.String()}
	}

	return []string{defaultAzureScope}
}

type azureSubscriptionMetadata struct {
	SubscriptionID string `json:"subscriptionId"`
}

// azurePricingsClient wraps armsecurity.PricingsClient with the subscription scope baked in
type azurePricingsClient struct {
	client *armsecurity.PricingsClient
	scope  string
}

// staticAzureCredential adapts a pre-obtained OAuth bearer token to azcore.TokenCredential
type staticAzureCredential struct {
	token string
}

// GetToken satisfies azcore.TokenCredential for a pre-obtained bearer token
func (s staticAzureCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// Provider implements client-credential authentication for Microsoft Defender for Cloud
type Provider struct {
	// BaseProvider provides shared provider metadata
	providers.BaseProvider
	tokenEndpoint func(tenantID types.TrimmedString) string
}

// Builder returns the Azure Security Center provider builder.
// The cfg parameter is reserved for future operator-level credential injection.
func Builder(_ Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeAzureSecurityCenter,
		SpecFunc:     azureSecurityCenterSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if err := providerkit.ValidateAuthType(s, types.AuthKindOAuth2ClientCredentials, ErrAuthTypeMismatch); err != nil {
				return nil, err
			}

			return newProvider(s), nil
		},
	}
}

// azureSecurityCenterSpec returns the static provider specification for Azure Security Center.
func azureSecurityCenterSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "azuresecuritycenter",
		DisplayName: "Microsoft Defender for Cloud",
		Category:    "compliance",
		AuthType:    types.AuthKindOAuth2ClientCredentials,
		Active:      lo.ToPtr(false),
		Visible:     lo.ToPtr(true),
		LogoURL:     "https://azure.microsoft.com/svghandler/microsoft-defender/?width=256&height=256",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center/overview",
		Labels: map[string]string{
			"vendor":  "microsoft",
			"product": "defender-for-cloud",
		},
		CredentialsSchema: azureSecurityCenterCredentialsSchema,
		Description:       "Collect Microsoft Defender for Cloud pricing and plan metadata from an Azure subscription for security posture visibility.",
	}
}

// newProvider constructs the Azure Security Center provider from a spec
func newProvider(s spec.ProviderSpec) *Provider {
	return &Provider{
		BaseProvider: providers.NewBaseProvider(
			TypeAzureSecurityCenter,
			types.ProviderCapabilities{
				SupportsRefreshTokens: true,
				SupportsClientPooling: true,
				SupportsMetadataForm:  len(s.CredentialsSchema) > 0,
			},
			providerkit.SanitizeOperationDescriptors(TypeAzureSecurityCenter, azureSecurityOperations()),
			providerkit.SanitizeClientDescriptors(TypeAzureSecurityCenter, azureSecurityCenterClientDescriptors()),
		),
		tokenEndpoint: defaultAzureTokenEndpoint,
	}
}

// BeginAuth is not supported for Azure Security Center client credentials
func (p *Provider) BeginAuth(context.Context, types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges stored client credentials for an Azure access token
func (p *Provider) Mint(ctx context.Context, subject types.CredentialMintRequest) (types.CredentialSet, error) {
	credentials, err := azureSecurityCenterMetadataFromPayload(subject.Credential)
	if err != nil {
		return types.CredentialSet{}, err
	}

	token, err := p.requestToken(ctx, credentials, subject.Scopes)
	if err != nil {
		return types.CredentialSet{}, err
	}

	providerData, err := jsonx.ToRawMessage(credentials.providerData())
	if err != nil {
		return types.CredentialSet{}, err
	}

	credential := types.CredentialSet{
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

// requestToken obtains an Azure access token using the client credentials flow
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

// defaultAzureTokenEndpoint builds the Azure token endpoint for the tenant
func defaultAzureTokenEndpoint(tenantID types.TrimmedString) string {
	if tenantID == "" {
		return ""
	}

	return fmt.Sprintf(azureTokenURLTemplate, tenantID)
}

// azureSecurityCenterMetadataFromPayload normalizes and validates provider metadata
func azureSecurityCenterMetadataFromPayload(payload types.CredentialSet) (azureSecurityCenterMetadata, error) {
	if len(payload.ProviderData) == 0 {
		return azureSecurityCenterMetadata{}, ErrProviderMetadataRequired
	}

	var decoded azureSecurityCenterMetadata
	if err := jsonx.UnmarshalIfPresent(payload.ProviderData, &decoded); err != nil {
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

// azureSecurityCenterClientDescriptors returns the client descriptors published by Defender for Cloud
func azureSecurityCenterClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeAzureSecurityCenter, ClientAzureSecurityCenterAPI, "Azure management API client for Defender for Cloud", buildAzureSecurityClient)
}

// newAzurePricingsClient constructs an azurePricingsClient from a bearer token and subscription ID
func newAzurePricingsClient(token string, subscriptionID string) (*azurePricingsClient, error) {
	if subscriptionID == "" {
		return nil, ErrSubscriptionIDMissing
	}

	client, err := armsecurity.NewPricingsClient(staticAzureCredential{token: token}, nil)
	if err != nil {
		return nil, err
	}

	return &azurePricingsClient{
		client: client,
		scope:  fmt.Sprintf("%s%s", azureSubscriptionScopePrefix, subscriptionID),
	}, nil
}

// buildAzureSecurityClient constructs an Azure Security Center client from a credential set
func buildAzureSecurityClient(_ context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token := credential.OAuthAccessToken
	if token == "" {
		return types.EmptyClientInstance(), providerkit.ErrOAuthTokenMissing
	}

	var meta azureSubscriptionMetadata
	if err := jsonx.UnmarshalIfPresent(credential.ProviderData, &meta); err != nil {
		return types.EmptyClientInstance(), err
	}

	apc, err := newAzurePricingsClient(token, meta.SubscriptionID)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(apc), nil
}

type azureSecurityHealthDetails struct {
	Count int `json:"count"`
}

type azureSecurityPricingSample struct {
	Name      string `json:"name"`
	Tier      string `json:"tier"`
	SubPlan   string `json:"subPlan"`
	FreeTrial string `json:"freeTrial"`
}

type azureSecurityPricingDetails struct {
	Count   int                          `json:"count"`
	Samples []azureSecurityPricingSample `json:"samples"`
}

// azureSecurityOperations registers the Defender for Cloud operations
func azureSecurityOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(azureSecurityHealth, "Call Azure Security Center pricings API to verify access.", ClientAzureSecurityCenterAPI, runAzureSecurityHealth),
		{
			Name:        azureSecurityPricing,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect plan/pricing metadata for Microsoft Defender for Cloud.",
			Client:      ClientAzureSecurityCenterAPI,
			Run:         runAzureSecurityPricing,
		},
	}
}

// resolveAzureSecurityClient returns a pooled Azure pricings client or builds one from the credential set
func resolveAzureSecurityClient(_ context.Context, input types.OperationInput) (*azurePricingsClient, error) {
	if c, ok := types.ClientInstanceAs[*azurePricingsClient](input.Client); ok {
		return c, nil
	}

	token := input.Credential.OAuthAccessToken
	if token == "" {
		return nil, providerkit.ErrOAuthTokenMissing
	}

	var meta azureSubscriptionMetadata
	if err := jsonx.UnmarshalIfPresent(input.Credential.ProviderData, &meta); err != nil {
		return nil, err
	}

	return newAzurePricingsClient(token, meta.SubscriptionID)
}

// runAzureSecurityHealth verifies access by fetching Defender pricing data
func runAzureSecurityHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	apc, err := resolveAzureSecurityClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := apc.client.List(ctx, apc.scope, nil)
	if err != nil {
		return providerkit.OperationFailure("Azure Security Center pricing fetch failed", err, nil)
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Retrieved %d pricing entries", len(resp.Value)), azureSecurityHealthDetails{
		Count: len(resp.Value),
	}), nil
}

// runAzureSecurityPricing collects Defender pricing metadata
func runAzureSecurityPricing(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	apc, err := resolveAzureSecurityClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := apc.client.List(ctx, apc.scope, nil)
	if err != nil {
		return providerkit.OperationFailure("Azure Security Center pricing fetch failed", err, nil)
	}

	sampleCount := min(len(resp.Value), providerkit.DefaultSampleSize)
	samples := lo.Map(resp.Value[:sampleCount], func(item *armsecurity.Pricing, _ int) azureSecurityPricingSample {
		sample := azureSecurityPricingSample{
			Name: lo.FromPtrOr(item.Name, ""),
		}

		if item.Properties != nil {
			if item.Properties.PricingTier != nil {
				sample.Tier = string(*item.Properties.PricingTier)
			}

			sample.SubPlan = lo.FromPtrOr(item.Properties.SubPlan, "")
			sample.FreeTrial = lo.FromPtrOr(item.Properties.FreeTrialRemainingTime, "")
		}

		return sample
	})

	return providerkit.OperationSuccess(fmt.Sprintf("Collected %d Defender pricing records", len(resp.Value)), azureSecurityPricingDetails{
		Count:   len(resp.Value),
		Samples: samples,
	}), nil
}
