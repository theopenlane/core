package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/user"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientCloudflareAPI identifies the Cloudflare HTTP API client
	ClientCloudflareAPI types.ClientName = "api"

	// TypeCloudflare identifies the Cloudflare provider
	TypeCloudflare types.ProviderType = "cloudflare"

	cloudflareHealthOp types.OperationName = types.OperationHealthDefault
)

type cloudflareHealthDetails struct {
	Status    string `json:"status,omitempty"`
	ExpiresOn string `json:"expiresOn,omitempty"`
}

// cloudflareCredentialsSchema is the JSON Schema for Cloudflare account credentials.
var cloudflareCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["apiToken","accountId"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this Cloudflare account."},"apiToken":{"type":"string","title":"API Token","description":"Scoped token with access to account settings, Zero Trust policies, and scanning features."},"accountId":{"type":"string","title":"Account ID","description":"Cloudflare account identifier targeted by this integration."},"zoneIds":{"type":"array","title":"Zone IDs","description":"Optional subset of zones to include when collecting findings.","items":{"type":"string"}},"email":{"type":"string","title":"Account Email","description":"Optional email associated with the API token (required for legacy key authentication)."},"apiKey":{"type":"string","title":"Global API Key","description":"Optional global API key when using email/key authentication instead of scoped tokens."},"enableScanner":{"type":"boolean","title":"Enable Scanner","description":"Toggle Cloudflare's scanning utility if the account is licensed for it.","default":false},"scannerLabel":{"type":"string","title":"Scanner Label","description":"Optional label recorded with scan jobs for traceability."}}}`)

// Builder returns the providers.Builder for the Cloudflare provider
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeCloudflare,
		SpecFunc:     cloudflareSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return apikey.Builder(
				TypeCloudflare,
				apikey.WithTokenField("apiToken"),
				apikey.WithClientDescriptors(cloudflareClientDescriptors()),
				apikey.WithOperations(cloudflareOperations()),
			).Build(ctx, s)
		},
	}
}

// cloudflareSpec returns the static provider specification for the Cloudflare provider.
func cloudflareSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "cloudflare",
		DisplayName: "Cloudflare",
		Category:    "security",
		AuthType:    types.AuthKindAPIKey,
		Active:      lo.ToPtr(true),
		Visible:     lo.ToPtr(true),
		LogoURL:     "https://developers.cloudflare.com/resources/logo/favicon.ico",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/cloudflare/overview",
		Labels: map[string]string{
			"vendor":  "cloudflare",
			"product": "zero-trust",
		},
		CredentialsSchema: cloudflareCredentialsSchema,
		Description:       "Validate Cloudflare account access and collect security-relevant account and zone context for posture workflows.",
	}
}

func cloudflareClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeCloudflare, ClientCloudflareAPI, "Cloudflare REST API client", buildCloudflareClient)
}

func buildCloudflareClient(_ context.Context, payload types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := providerkit.APITokenFromCredential(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(cf.NewClient(option.WithAPIToken(token))), nil
}

func cloudflareOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(cloudflareHealthOp, "Verify Cloudflare API token via /user/tokens/verify.", ClientCloudflareAPI, runCloudflareHealth),
	}
}

func resolveCloudflareClient(input types.OperationInput) (*cf.Client, error) {
	if c, ok := types.ClientInstanceAs[*cf.Client](input.Client); ok {
		return c, nil
	}

	token, err := providerkit.APITokenFromCredential(input.Credential)
	if err != nil {
		return nil, err
	}

	return cf.NewClient(option.WithAPIToken(token)), nil
}

func runCloudflareHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveCloudflareClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	res, err := client.User.Tokens.Verify(ctx)
	if err != nil {
		return providerkit.OperationFailure("Cloudflare token verification failed", err, nil)
	}

	if res.Status != user.TokenVerifyResponseStatusActive {
		return providerkit.OperationFailure("Cloudflare token is not active", ErrTokenVerificationFailed, cloudflareHealthDetails{
			Status: string(res.Status),
		})
	}

	details := cloudflareHealthDetails{
		Status: string(res.Status),
	}

	if !res.ExpiresOn.IsZero() {
		details.ExpiresOn = res.ExpiresOn.String()
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Cloudflare token verified, status: %s", res.Status), details), nil
}
