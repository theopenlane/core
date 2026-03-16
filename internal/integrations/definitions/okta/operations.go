package okta

import (
	"context"
	"encoding/json"
	"fmt"

	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	oktaSignOnPolicyType = "OKTA_SIGN_ON"
	sampleSize           = 10
)

// HealthCheck holds the result of an Okta health check
type HealthCheck struct {
	// ID is the Okta user identifier
	ID string `json:"id"`
	// Login is the Okta user login
	Login string `json:"login"`
	// Email is the Okta user email
	Email string `json:"email"`
}

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

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Okta health check
func (HealthCheck) Run(ctx context.Context, c *oktagosdk.APIClient) (json.RawMessage, error) {
	user, _, err := c.UserAPI.GetUser(ctx, "me").Execute()
	if err != nil {
		return nil, fmt.Errorf("okta: user lookup failed: %w", err)
	}

	profile := user.GetProfile()
	login := profile.GetLogin()

	return jsonx.ToRawMessage(HealthCheck{
		ID:    user.GetId(),
		Login: login,
		Email: profile.GetEmail(),
	})
}

// Handle adapts policy collection to the generic operation registration boundary
func (p PoliciesCollect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return p.Run(ctx, c)
	}
}

// Run collects Okta sign-on policy metadata
func (PoliciesCollect) Run(ctx context.Context, c *oktagosdk.APIClient) (json.RawMessage, error) {
	policies, _, err := c.PolicyAPI.ListPolicies(ctx).Type_(oktaSignOnPolicyType).Execute()
	if err != nil {
		return nil, fmt.Errorf("okta: policies fetch failed: %w", err)
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

	return jsonx.ToRawMessage(PoliciesCollect{
		Count:   len(policies),
		Samples: samples,
	})
}
