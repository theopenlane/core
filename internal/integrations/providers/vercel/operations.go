package vercel

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	vercelHealthOp   types.OperationName = "health.default"
	vercelProjectsOp types.OperationName = "projects.sample"
)

// vercelOperations returns the Vercel operations supported by this provider.
func vercelOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        vercelHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Vercel /v2/user to verify token and account.",
			Client:      ClientVercelAPI,
			Run:         runVercelHealth,
		},
		{
			Name:        vercelProjectsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of Vercel projects for drift detection.",
			Client:      ClientVercelAPI,
			Run:         runVercelProjects,
		},
	}
}

func runVercelHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeVercel))
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		User struct {
			ID    string `json:"uid"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}

	endpoint := "https://api.vercel.com/v2/user"
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &resp); err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Vercel user lookup failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
	} else if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Vercel user lookup failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.APITokenFromPayload(input.Credential, string(TypeVercel))
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("limit", "5")

	var resp struct {
		Projects []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Framework string `json:"framework"`
		} `json:"projects"`
	}

	endpoint := "https://api.vercel.com/v4/projects?" + params.Encode()

	if client != nil {
		err = client.GetJSON(ctx, endpoint, &resp)
	} else {
		err = helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp)
	}
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Vercel projects fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
