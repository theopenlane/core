package buildkite

import (
	"context"
	"errors"
	"fmt"

	"github.com/theopenlane/common/integrations/helpers"
	"github.com/theopenlane/common/integrations/types"
)

const (
	buildkiteOperationHealth types.OperationName = "health.default"
	buildkiteOperationOrgs   types.OperationName = "organizations.collect"

	maxSampleSize = 5
)

func buildkiteOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        buildkiteOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Validate Buildkite token by calling the /v2/user endpoint.",
			Run:         runBuildkiteHealthOperation,
		},
		{
			Name:        buildkiteOperationOrgs,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Buildkite organizations for reporting.",
			Run:         runBuildkiteOrganizationsOperation,
		},
	}
}

type buildkiteUserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type buildkiteOrgResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	WebURL string `json:"web_url"`
}

func runBuildkiteHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeBuildkite))
	if err != nil {
		return types.OperationResult{}, err
	}

	var user buildkiteUserResponse
	if err := fetchBuildkiteResource(ctx, token, "user", &user); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Buildkite user lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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

func runBuildkiteOrganizationsOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeBuildkite))
	if err != nil {
		return types.OperationResult{}, err
	}

	var orgs []buildkiteOrgResponse
	if err := fetchBuildkiteResource(ctx, token, "organizations", &orgs); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Buildkite organizations fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	samples := make([]map[string]any, 0, minInt(maxSampleSize, len(orgs)))
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

func fetchBuildkiteResource(ctx context.Context, token, path string, out any) error {
	endpoint := fmt.Sprintf("https://api.buildkite.com/v2/%s", path)
	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, out); err != nil {
		if errors.Is(err, helpers.ErrHTTPRequestFailed) {
			return fmt.Errorf("%w (path %s): %s", ErrAPIRequest, path, err.Error())
		}
		return err
	}

	return nil
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
