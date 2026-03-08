package vercel

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	vercelHealthOp   types.OperationName = types.OperationHealthDefault
	vercelProjectsOp types.OperationName = "projects.sample"
	vercelAPIBaseURL                     = "https://api.vercel.com"
)

type vercelUserResponse struct {
	User struct {
		ID    string `json:"uid"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
}

type vercelHealthDetails struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type vercelProjectSample struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework"`
}

type vercelProjectsDetails struct {
	Projects []vercelProjectSample `json:"projects"`
}

// vercelOperations returns the Vercel operations supported by this provider.
func vercelOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(vercelHealthOp, "Call Vercel /v2/user to verify token and account.", ClientVercelAPI,
			operations.HealthCheckRunner(auth.APITokenFromPayload, "https://api.vercel.com/v2/user", "Vercel user lookup failed",
				func(resp vercelUserResponse) (string, any) {
					return fmt.Sprintf("Vercel token valid for %s", resp.User.Email), vercelHealthDetails{
						ID:    resp.User.ID,
						Name:  resp.User.Name,
						Email: resp.User.Email,
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
	client, err := auth.ResolveAuthenticatedClient(input, auth.APITokenFromPayload, vercelAPIBaseURL, nil)
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

	if err := client.GetJSONWithParams(ctx, "/v4/projects", params, &resp); err != nil {
		return operations.OperationFailure("Vercel projects fetch failed", err, nil)
	}

	samples := make([]vercelProjectSample, 0, len(resp.Projects))
	for _, project := range resp.Projects {
		samples = append(samples, vercelProjectSample{
			ID:        project.ID,
			Name:      project.Name,
			Framework: project.Framework,
		})
	}

	return operations.OperationSuccess(fmt.Sprintf("Fetched %d Vercel projects", len(samples)), vercelProjectsDetails{Projects: samples}), nil
}
