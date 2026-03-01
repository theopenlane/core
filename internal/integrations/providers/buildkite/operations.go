package buildkite

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	buildkiteOperationHealth types.OperationName = "health.default"
	buildkiteOperationOrgs   types.OperationName = "organizations.collect"
)

// buildkiteUserResponse represents the response from the Buildkite /v2/user endpoint
type buildkiteUserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// buildkiteOrgResponse represents the response from the Buildkite /v2/organizations endpoint
type buildkiteOrgResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	WebURL string `json:"web_url"`
}

// buildkiteOperations returns the Buildkite operations supported by this provider.
func buildkiteOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(buildkiteOperationHealth, "Validate Buildkite token by calling the /v2/user endpoint.", ClientBuildkiteAPI,
			operations.HealthCheckRunner(operations.TokenTypeAPI, "https://api.buildkite.com/v2/user", "Buildkite user lookup failed",
				func(user buildkiteUserResponse) (string, map[string]any) {
					return fmt.Sprintf("Buildkite token valid for %s", user.Name), map[string]any{
						"id":       user.ID,
						"name":     user.Name,
						"email":    user.Email,
						"username": user.Username,
					}
				})),
		{
			Name:        buildkiteOperationOrgs,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Buildkite organizations for reporting.",
			Client:      ClientBuildkiteAPI,
			Run:         runBuildkiteOrganizationsOperation,
		},
	}
}

// runBuildkiteOrganizationsOperation collects Buildkite org metadata for reporting.
func runBuildkiteOrganizationsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndToken(input, auth.APITokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	var orgs []buildkiteOrgResponse
	endpoint := "https://api.buildkite.com/v2/organizations"
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &orgs); err != nil {
		return operations.OperationFailure("Buildkite organizations fetch failed", err, nil)
	}

	samples := lo.Map(orgs[:min(len(orgs), operations.DefaultSampleSize)], func(org buildkiteOrgResponse, _ int) map[string]any {
		return map[string]any{
			"id":   org.ID,
			"name": org.Name,
			"slug": org.Slug,
			"url":  org.WebURL,
		}
	})

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Discovered %d Buildkite organizations", len(orgs)),
		Details: map[string]any{
			"count":   len(orgs),
			"samples": samples,
		},
	}, nil
}
