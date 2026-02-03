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
		{
			Name:        googleWorkspaceHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Google OAuth userinfo to verify the workspace token.",
			Client:      ClientGoogleWorkspaceAPI,
			Run:         runGoogleWorkspaceHealth,
		},
		{
			Name:        googleWorkspaceUsersOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "List sample Admin Directory users for posture checks.",
			Client:      ClientGoogleWorkspaceAPI,
			Run:         runGoogleWorkspaceUsers,
		},
	}
}

func runGoogleWorkspaceHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeGoogleWorkspace))
	if err != nil {
		return types.OperationResult{}, err
	}

	var userinfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	endpoint := "https://www.googleapis.com/oauth2/v3/userinfo"
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &userinfo); err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Google userinfo failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
	} else if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &userinfo); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Google userinfo failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeGoogleWorkspace))
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

	if client != nil {
		err = client.GetJSON(ctx, endpoint, &resp)
	} else {
		err = helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp)
	}
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Directory users fetch failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
