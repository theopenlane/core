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

type vercelUserResponse struct {
	User struct {
		ID    string `json:"uid"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

// vercelOperations returns the Vercel operations supported by this provider.
func vercelOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(vercelHealthOp, "Call Vercel /v2/user to verify token and account.", ClientVercelAPI,
			operations.HealthCheckRunner(operations.TokenTypeAPI, "https://api.vercel.com/v2/user", "Vercel user lookup failed",
				func(resp vercelUserResponse) (string, map[string]any) {
					return fmt.Sprintf("Vercel token valid for %s", resp.User.Email), map[string]any{
						"id":    resp.User.ID,
						"name":  resp.User.Name,
						"email": resp.User.Email,
					}
				})),
		{
			Name:        vercelProjectsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of Vercel projects for drift detection.",
			Client:      ClientVercelAPI,
			Run:         runVercelProjects,
		},
	}
}

// runVercelProjects returns a sample of projects for drift detection.
func runVercelProjects(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndAPIToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("limit", "5")

	var resp struct {
		// Projects lists projects returned by the API
		Projects []struct {
			// ID is the project identifier
			ID string `json:"id"`
			// Name is the project name
			Name string `json:"name"`
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
