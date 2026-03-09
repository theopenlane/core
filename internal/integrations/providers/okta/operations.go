package okta

import (
	"context"
	"fmt"

	okta "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	oktaHealthOp   types.OperationName = types.OperationHealthDefault
	oktaPoliciesOp types.OperationName = "policies.collect"

	oktaSignOnPolicyType = "OKTA_SIGN_ON"
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

// oktaOperations returns the Okta operations supported by this provider.
func oktaOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(oktaHealthOp, "Call Okta user API to verify API token.", ClientOktaAPI, runOktaHealth),
		{
			Name:        oktaPoliciesOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect sign-on policy metadata for posture analysis.",
			Client:      ClientOktaAPI,
			Run:         runOktaPolicies,
		},
	}
}

// resolveOktaClient returns a pooled Okta client or builds one from the credential payload.
func resolveOktaClient(input types.OperationInput) (*okta.APIClient, error) {
	if c, ok := types.ClientInstanceAs[*okta.APIClient](input.Client); ok {
		return c, nil
	}

	apiToken, err := auth.APITokenFromPayload(input.Credential)
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

// runOktaHealth verifies the Okta API token by fetching the current user.
func runOktaHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveOktaClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	user, _, err := client.UserAPI.GetUser(ctx, "me").Execute()
	if err != nil {
		return operations.OperationFailure("Okta user lookup failed", err, nil)
	}

	profile := user.GetProfile()
	login := profile.GetLogin()

	return operations.OperationSuccess(fmt.Sprintf("Okta token valid for %s", login), oktaHealthDetails{
		ID:    user.GetId(),
		Login: login,
		Email: profile.GetEmail(),
	}), nil
}

// runOktaPolicies collects a sample of Okta sign-on policies for reporting.
func runOktaPolicies(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveOktaClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	policies, _, err := client.PolicyAPI.ListPolicies(ctx).Type_(oktaSignOnPolicyType).Execute()
	if err != nil {
		return operations.OperationFailure("Okta policies fetch failed", err, nil)
	}

	samples := lo.Map(policies[:min(len(policies), operations.DefaultSampleSize)], func(item okta.ListPolicies200ResponseInner, _ int) oktaPolicySample {
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

	return operations.OperationSuccess(fmt.Sprintf("Collected %d sign-on policies", len(policies)), oktaPoliciesDetails{
		Count:   len(policies),
		Samples: samples,
	}), nil
}
