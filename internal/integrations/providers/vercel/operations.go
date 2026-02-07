package vercel

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	vercelHealthOp   types.OperationName = "health.default"
	vercelProjectsOp types.OperationName = "projects.sample"
)

// vercelOperations returns the Vercel operations supported by this provider.
func vercelOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(vercelHealthOp, "Call Vercel /v2/user to verify token and account.", ClientVercelAPI, runVercelHealth),
		{
			Name:        vercelProjectsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of Vercel projects for drift detection.",
			Client:      ClientVercelAPI,
			Run:         runVercelProjects,
		},
	}
}

// runVercelHealth verifies the Vercel API token by fetching the user account
func runVercelHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndAPIToken(input, TypeVercel)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		// User holds the authenticated user account
		User struct {
			// ID is the user identifier
			ID    string `json:"uid"`
			// Name is the user display name
			Name  string `json:"name"`
			// Email is the user email address
			Email string `json:"email"`
		} `json:"user"`
	}

	endpoint := "https://api.vercel.com/v2/user"
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return operations.OperationFailure("Vercel user lookup failed", err), err
	}

	summary := fmt.Sprintf("Vercel token valid for %s", resp.User.Email)
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: map[string]any{
			"id":    resp.User.ID,
			"name":  resp.User.Name,
			"email": resp.User.Email,
		},
	}, nil
}

// runVercelProjects returns a sample of projects for drift detection.
func runVercelProjects(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndAPIToken(input, TypeVercel)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("limit", "5")

	var resp struct {
		// Projects lists projects returned by the API
		Projects []struct {
			// ID is the project identifier
			ID        string `json:"id"`
			// Name is the project name
			Name      string `json:"name"`
			// Framework is the detected framework name
			Framework string `json:"framework"`
		} `json:"projects"`
	}

	endpoint := "https://api.vercel.com/v4/projects?" + params.Encode()

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return operations.OperationFailure("Vercel projects fetch failed", err), err
	}

	samples := make([]map[string]any, 0, len(resp.Projects))
	for _, project := range resp.Projects {
		samples = append(samples, map[string]any{
			"id":        project.ID,
			"name":      project.Name,
			"framework": project.Framework,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Fetched %d Vercel projects", len(samples)),
		Details: map[string]any{"projects": samples},
	}, nil
}
