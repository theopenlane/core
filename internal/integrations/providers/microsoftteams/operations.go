package microsoftteams

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	teamsHealthOp   types.OperationName = "health.default"
	teamsChannelsOp types.OperationName = "teams.sample"
)

// teamsOperations returns the Microsoft Teams operations supported by this provider
func teamsOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(teamsHealthOp, "Call Graph /me to verify Teams access.", ClientMicrosoftTeamsAPI, runTeamsHealth),
		{
			Name:        teamsChannelsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of joined teams for the user context.",
			Client:      ClientMicrosoftTeamsAPI,
			Run:         runTeamsSample,
		},
	}
}

// runTeamsHealth verifies the Microsoft Graph token by fetching the user profile
func runTeamsHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndOAuthToken(input, TypeMicrosoftTeams)
	if err != nil {
		return types.OperationResult{}, err
	}

	var profile struct {
		// ID is the user identifier
		ID          string `json:"id"`
		// DisplayName is the user display name
		DisplayName string `json:"displayName"`
		// Mail is the primary email address
		Mail        string `json:"mail"`
	}

	endpoint := "https://graph.microsoft.com/v1.0/me"
	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &profile); err != nil {
		return operations.OperationFailure("Graph /me failed", err), err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Graph token valid for %s", profile.DisplayName),
		Details: map[string]any{"id": profile.ID, "mail": profile.Mail},
	}, nil
}

// runTeamsSample collects a sample of joined Teams for the authenticated user
func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndOAuthToken(input, TypeMicrosoftTeams)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		// Value lists the joined teams
		Value []struct {
			// ID is the team identifier
			ID          string `json:"id"`
			// DisplayName is the team display name
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	endpoint := "https://graph.microsoft.com/v1.0/me/joinedTeams?$top=5"

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return operations.OperationFailure("Graph joinedTeams failed", err), err
	}

	samples := make([]map[string]any, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, map[string]any{
			"id":          team.ID,
			"displayName": team.DisplayName,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Retrieved %d joined teams", len(samples)),
		Details: map[string]any{"teams": samples},
	}, nil
}
