package buildkite

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	buildkiteOperationHealth types.OperationName = "health.default"
	buildkiteOperationOrgs   types.OperationName = "organizations.collect"

	maxSampleSize = 5
)

// buildkiteOperations returns the Buildkite operations supported by this provider.
func buildkiteOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(buildkiteOperationHealth, "Validate Buildkite token by calling the /v2/user endpoint.", ClientBuildkiteAPI, runBuildkiteHealthOperation),
		{
			Name:        buildkiteOperationOrgs,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Buildkite organizations for reporting.",
			Client:      ClientBuildkiteAPI,
			Run:         runBuildkiteOrganizationsOperation,
		},
	}
}

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

// runBuildkiteHealthOperation verifies access to the Buildkite API using the current token.
func runBuildkiteHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndAPIToken(input, TypeBuildkite)
	if err != nil {
		return types.OperationResult{}, err
	}

	var user buildkiteUserResponse
	if err := fetchBuildkiteResource(ctx, client, token, "user", &user); err != nil {
		return operations.OperationFailure("Buildkite user lookup failed", err), err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Buildkite token valid for %s", user.Name),
		Details: map[string]any{
			"id":       user.ID,
			"name":     user.Name,
			"email":    user.Email,
			"username": user.Username,
		},
	}, nil
}

// runBuildkiteOrganizationsOperation collects Buildkite org metadata for reporting.
func runBuildkiteOrganizationsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndAPIToken(input, TypeBuildkite)
	if err != nil {
		return types.OperationResult{}, err
	}

	var orgs []buildkiteOrgResponse
	if err := fetchBuildkiteResource(ctx, client, token, "organizations", &orgs); err != nil {
		return operations.OperationFailure("Buildkite organizations fetch failed", err), err
	}

	samples := make([]map[string]any, 0, min(maxSampleSize, len(orgs)))
	for _, org := range orgs {
		if len(samples) >= cap(samples) {
			break
		}
		samples = append(samples, map[string]any{
			"id":   org.ID,
			"name": org.Name,
			"slug": org.Slug,
			"url":  org.WebURL,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Discovered %d Buildkite organizations", len(orgs)),
		Details: map[string]any{
			"count":   len(orgs),
			"samples": samples,
		},
	}, nil
}

// fetchBuildkiteResource retrieves Buildkite API resources with either a pooled client or raw token.
func fetchBuildkiteResource(ctx context.Context, client *auth.AuthenticatedClient, token, path string, out any) error {
	endpoint := fmt.Sprintf("https://api.buildkite.com/v2/%s", path)
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, out); err != nil {
		if errors.Is(err, auth.ErrHTTPRequestFailed) {
			return fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}
		return err
	}

	return nil
}
