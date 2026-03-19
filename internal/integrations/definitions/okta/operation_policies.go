package okta

import (
	"context"
	"encoding/json"

	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// oktaSignOnPolicyType is the Okta policy type identifier for sign-on policies
	oktaSignOnPolicyType = "OKTA_SIGN_ON"
	// sampleSize is the maximum number of policies to include in the sample result
	sampleSize = 10
)

// PolicySample holds a single Okta policy entry
type PolicySample struct {
	// ID is the Okta policy identifier
	ID string `json:"id"`
	// Name is the policy display name
	Name string `json:"name"`
	// Status is the current policy status
	Status string `json:"status"`
	// Type is the policy type
	Type string `json:"type"`
}

// PoliciesCollect collects Okta sign-on policy metadata
type PoliciesCollect struct {
	// Count is the total number of policies collected
	Count int `json:"count"`
	// Samples holds a representative subset of collected policies
	Samples []PolicySample `json:"samples"`
}

// Handle adapts policy collection to the generic operation registration boundary
func (p PoliciesCollect) Handle() types.OperationHandler {
	return providerkit.OperationWithClient(OktaClient, p.Run)
}

// Run collects Okta sign-on policy metadata
func (PoliciesCollect) Run(ctx context.Context, c *oktagosdk.APIClient) (json.RawMessage, error) {
	policies, _, err := c.PolicyAPI.ListPolicies(ctx).Type_(oktaSignOnPolicyType).Execute()
	if err != nil {
		return nil, ErrPoliciesFetchFailed
	}

	samples := lo.Map(policies[:min(len(policies), sampleSize)], func(item oktagosdk.ListPolicies200ResponseInner, _ int) PolicySample {
		if p := item.OktaSignOnPolicy; p != nil {
			return PolicySample{
				ID:     p.GetId(),
				Name:   p.GetName(),
				Status: p.GetStatus(),
				Type:   p.GetType(),
			}
		}

		return PolicySample{}
	})

	return providerkit.EncodeResult(PoliciesCollect{
		Count:   len(policies),
		Samples: samples,
	}, ErrResultEncode)
}
