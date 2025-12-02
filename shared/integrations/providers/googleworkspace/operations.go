package googleworkspace

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/shared/integrations/providers/helpers"
	"github.com/theopenlane/shared/integrations/types"
)

const (
	googleWorkspaceHealthOp types.OperationName = "health.default"
	googleWorkspaceUsersOp  types.OperationName = "directory.sample_users"
)

func googleWorkspaceOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        googleWorkspaceHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Google OAuth userinfo to verify the workspace token.",
			Run:         runGoogleWorkspaceHealth,
		},
		{
			Name:        googleWorkspaceUsersOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "List sample Admin Directory users for posture checks.",
			Run:         runGoogleWorkspaceUsers,
		},
	}
}

func runGoogleWorkspaceHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeGoogleWorkspace))
	if err != nil {
		return types.OperationResult{}, err
	}

	var userinfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := helpers.HTTPGetJSON(ctx, nil, "https://www.googleapis.com/oauth2/v3/userinfo", token, nil, &userinfo); err != nil {
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

func runGoogleWorkspaceUsers(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
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

	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp); err != nil {
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
