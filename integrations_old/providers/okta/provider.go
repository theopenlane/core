package okta

import (
	"context"
	"encoding/json"
	"fmt"

	okta "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/apikey"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// ClientOktaAPI identifies the Okta HTTP API client
	ClientOktaAPI types.ClientName = "api"

	// TypeOkta identifies the Okta provider
	TypeOkta types.ProviderType = "okta"

	oktaHealthOp   types.OperationName = types.OperationHealthDefault
	oktaPoliciesOp types.OperationName = "policies.collect"

	oktaSignOnPolicyType = "OKTA_SIGN_ON"
)

type oktaProviderMetadata struct {
	OrgURL string `json:"orgUrl"`
}

type oktaHealthDetails struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Email string `json:"email"`
}

type oktaPolicySample struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

type oktaPoliciesDetails struct {
	Count   int                `json:"count"`
	Samples []oktaPolicySample `json:"samples"`
}

// oktaCredentialsSchema is the JSON Schema for Okta tenant credentials.
var oktaCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["orgUrl","apiToken"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this Okta tenant."},"orgUrl":{"type":"string","title":"Org URL","description":"Base URL of the Okta organization (e.g. https://acme.okta.com)."},"apiToken":{"type":"string","title":"API Token","description":"Okta API token with permission to query users, groups, and system settings."}}}`)

// Builder returns the providers.Builder for the Okta provider
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeOkta,
		SpecFunc:     oktaSpec,
		BuildFunc: func(ctx context.Context, s spec.ProviderSpec) (types.Provider, error) {
			return apikey.Builder(
				TypeOkta,
				apikey.WithTokenField("apiToken"),
				apikey.WithClientDescriptors(oktaClientDescriptors()),
				apikey.WithOperations(oktaOperations()),
			).Build(ctx, s)
		},
	}
}

// oktaSpec returns the static provider specification for the Okta provider.
func oktaSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:        "okta",
		DisplayName: "Okta",
		Category:    "sso",
		AuthType:    types.AuthKindAPIKey,
		Active:      lo.ToPtr(false),
		Visible:     lo.ToPtr(true),
		LogoURL:     "",
		DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/okta/overview",
		Labels: map[string]string{
			"vendor":  "okta",
			"product": "identity",
		},
		CredentialsSchema: oktaCredentialsSchema,
		Description:       "Collect Okta tenant and sign-on policy metadata for identity posture and access governance analysis.",
	}
}

func oktaClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeOkta, ClientOktaAPI, "Okta API client", buildOktaClient)
}

func buildOktaClient(_ context.Context, payload types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	apiToken, err := providerkit.APITokenFromCredential(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	var meta oktaProviderMetadata
	if err := jsonx.UnmarshalIfPresent(payload.ProviderData, &meta); err != nil {
		return types.EmptyClientInstance(), err
	}

	orgURL := meta.OrgURL
	if orgURL == "" {
		return types.EmptyClientInstance(), ErrCredentialsMissing
	}

	cfg, err := okta.NewConfiguration(
		okta.WithOrgUrl(orgURL),
		okta.WithToken(apiToken),
	)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(okta.NewAPIClient(cfg)), nil
}

func oktaOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(oktaHealthOp, "Call Okta user API to verify API token.", ClientOktaAPI, runOktaHealth),
		{
			Name:        oktaPoliciesOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect sign-on policy metadata for posture analysis.",
			Client:      ClientOktaAPI,
			Run:         runOktaPolicies,
		},
	}
}

func resolveOktaClient(input types.OperationInput) (*okta.APIClient, error) {
	if c, ok := types.ClientInstanceAs[*okta.APIClient](input.Client); ok {
		return c, nil
	}

	apiToken, err := providerkit.APITokenFromCredential(input.Credential)
	if err != nil {
		return nil, err
	}

	var meta oktaProviderMetadata
	if err := jsonx.UnmarshalIfPresent(input.Credential.ProviderData, &meta); err != nil {
		return nil, err
	}

	orgURL := meta.OrgURL
	if orgURL == "" {
		return nil, ErrCredentialsMissing
	}

	cfg, err := okta.NewConfiguration(
		okta.WithOrgUrl(orgURL),
		okta.WithToken(apiToken),
	)
	if err != nil {
		return nil, err
	}

	return okta.NewAPIClient(cfg), nil
}

func runOktaHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveOktaClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	user, _, err := client.UserAPI.GetUser(ctx, "me").Execute()
	if err != nil {
		return providerkit.OperationFailure("Okta user lookup failed", err, nil)
	}

	profile := user.GetProfile()
	login := profile.GetLogin()

	return providerkit.OperationSuccess(fmt.Sprintf("Okta token valid for %s", login), oktaHealthDetails{
		ID:    user.GetId(),
		Login: login,
		Email: profile.GetEmail(),
	}), nil
}

func runOktaPolicies(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveOktaClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	policies, _, err := client.PolicyAPI.ListPolicies(ctx).Type_(oktaSignOnPolicyType).Execute()
	if err != nil {
		return providerkit.OperationFailure("Okta policies fetch failed", err, nil)
	}

	samples := lo.Map(policies[:min(len(policies), providerkit.DefaultSampleSize)], func(item okta.ListPolicies200ResponseInner, _ int) oktaPolicySample {
		if p := item.OktaSignOnPolicy; p != nil {
			return oktaPolicySample{
				ID:     p.GetId(),
				Name:   p.GetName(),
				Status: p.GetStatus(),
				Type:   p.GetType(),
			}
		}

		return oktaPolicySample{}
	})

	return providerkit.OperationSuccess(fmt.Sprintf("Collected %d sign-on policies", len(policies)), oktaPoliciesDetails{
		Count:   len(policies),
		Samples: samples,
	}), nil
}
