package okta

import (
	"context"
	"encoding/json"
	"fmt"

	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	oktaSignOnPolicyType = "OKTA_SIGN_ON"
	sampleSize           = 10
)

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

// buildOktaClient builds the Okta API client for one installation
func buildOktaClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	if cred.OrgURL == "" {
		return nil, ErrOrgURLMissing
	}

	cfg, err := oktagosdk.NewConfiguration(
		oktagosdk.WithOrgUrl(cred.OrgURL),
		oktagosdk.WithToken(cred.APIToken),
	)
	if err != nil {
		return nil, err
	}

	return oktagosdk.NewAPIClient(cfg), nil
}

// runHealthOperation validates the Okta API token by calling the user endpoint
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	oktaClient, ok := client.(*oktagosdk.APIClient)
	if !ok {
		return nil, ErrClientType
	}

	user, _, err := oktaClient.UserAPI.GetUser(ctx, "me").Execute()
	if err != nil {
		return nil, fmt.Errorf("okta: user lookup failed: %w", err)
	}

	profile := user.GetProfile()
	login := profile.GetLogin()

	return jsonx.ToRawMessage(oktaHealthDetails{
		ID:    user.GetId(),
		Login: login,
		Email: profile.GetEmail(),
	})
}

// runPoliciesCollectOperation collects Okta sign-on policy metadata
func runPoliciesCollectOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	oktaClient, ok := client.(*oktagosdk.APIClient)
	if !ok {
		return nil, ErrClientType
	}

	policies, _, err := oktaClient.PolicyAPI.ListPolicies(ctx).Type_(oktaSignOnPolicyType).Execute()
	if err != nil {
		return nil, fmt.Errorf("okta: policies fetch failed: %w", err)
	}

	samples := lo.Map(policies[:min(len(policies), sampleSize)], func(item oktagosdk.ListPolicies200ResponseInner, _ int) oktaPolicySample {
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

	return jsonx.ToRawMessage(oktaPoliciesDetails{
		Count:   len(policies),
		Samples: samples,
	})
}
