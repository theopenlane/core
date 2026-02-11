package googleworkspace

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	googleWorkspaceHealthOp types.OperationName = "health.default"
	googleWorkspaceUsersOp  types.OperationName = "directory.sample_users"
)

// googleWorkspaceOperations returns the Google Workspace operations supported by this provider.
func googleWorkspaceOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(googleWorkspaceHealthOp, "Call Google OAuth userinfo to verify the workspace token.", ClientGoogleWorkspaceAPI, runGoogleWorkspaceHealth),
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
	client, token, err := auth.ClientAndOAuthToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	var userinfo struct {
		// Sub is the subject identifier for the user
		Sub string `json:"sub"`
		// Email is the primary email address
		Email string `json:"email"`
		// Name is the display name for the user
		Name string `json:"name"`
	}

	endpoint := "https://www.googleapis.com/oauth2/v3/userinfo"
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &userinfo); err != nil {
		return operations.OperationFailure("Google userinfo failed", err), err
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
	client, token, err := auth.ClientAndOAuthToken(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("customer", "my_customer")
	params.Set("maxResults", "5")

	endpoint := "https://admin.googleapis.com/admin/directory/v1/users?" + params.Encode()
	var resp struct {
		// Users lists users returned from the directory API
		Users []struct {
			// PrimaryEmail is the user's primary email address
			PrimaryEmail string `json:"primaryEmail"`
			// Name holds the user's name metadata
			Name struct {
				// FullName is the user's full display name
				FullName string `json:"fullName"`
			} `json:"name"`
		} `json:"users"`
	}

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return operations.OperationFailure("Directory users fetch failed", err), err
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
