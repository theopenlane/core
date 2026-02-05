package googleworkspace

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	googleWorkspaceHealthOp types.OperationName = "health.default"
	googleWorkspaceUsersOp  types.OperationName = "directory.sample_users"
)

// googleWorkspaceOperations returns the Google Workspace operations supported by this provider.
func googleWorkspaceOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		helpers.HealthOperation(googleWorkspaceHealthOp, "Call Google OAuth userinfo to verify the workspace token.", ClientGoogleWorkspaceAPI, runGoogleWorkspaceHealth),
		{
			Name:        googleWorkspaceUsersOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "List sample Admin Directory users for posture checks.",
			Client:      ClientGoogleWorkspaceAPI,
			Run:         runGoogleWorkspaceUsers,
		},
	}
}

// runGoogleWorkspaceHealth validates the OAuth token with userinfo
func runGoogleWorkspaceHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, TypeGoogleWorkspace)
	if err != nil {
		return types.OperationResult{}, err
	}

	var userinfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	endpoint := "https://www.googleapis.com/oauth2/v3/userinfo"
	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, nil, &userinfo); err != nil {
		return helpers.OperationFailure("Google userinfo failed", err), err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Google token valid for %s", userinfo.Email),
		Details: map[string]any{
			"sub":   userinfo.Sub,
			"email": userinfo.Email,
			"name":  userinfo.Name,
		},
	}, nil
}

// runGoogleWorkspaceUsers returns a small sample of directory users for posture checks.
func runGoogleWorkspaceUsers(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, TypeGoogleWorkspace)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("customer", "my_customer")
	params.Set("maxResults", "5")

	endpoint := "https://admin.googleapis.com/admin/directory/v1/users?" + params.Encode()
	var resp struct {
		Users []struct {
			PrimaryEmail string `json:"primaryEmail"`
			Name         struct {
				FullName string `json:"fullName"`
			} `json:"name"`
		} `json:"users"`
	}

	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return helpers.OperationFailure("Directory users fetch failed", err), err
	}

	samples := make([]map[string]any, 0, len(resp.Users))
	for _, user := range resp.Users {
		samples = append(samples, map[string]any{
			"email": user.PrimaryEmail,
			"name":  user.Name.FullName,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Fetched %d sample users", len(samples)),
		Details: map[string]any{
			"samples": samples,
		},
	}, nil
}
