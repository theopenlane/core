package microsoftteams

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	teamsHealthOp   types.OperationName = "health.default"
	teamsChannelsOp types.OperationName = "teams.sample"
)

// teamsOperations returns the Microsoft Teams operations supported by this provider
func teamsOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        teamsHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Graph /me to verify Teams access.",
			Client:      ClientMicrosoftTeamsAPI,
			Run:         runTeamsHealth,
		},
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
	client, token, err := helpers.ClientAndOAuthToken(input, TypeMicrosoftTeams)
	if err != nil {
		return types.OperationResult{}, err
	}

	var profile struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
	}

	endpoint := "https://graph.microsoft.com/v1.0/me"
	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, nil, &profile); err != nil {
		return helpers.OperationFailure("Graph /me failed", err), err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Graph token valid for %s", profile.DisplayName),
		Details: map[string]any{"id": profile.ID, "mail": profile.Mail},
	}, nil
}

// runTeamsSample collects a sample of joined Teams for the authenticated user
func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, TypeMicrosoftTeams)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	endpoint := "https://graph.microsoft.com/v1.0/me/joinedTeams?$top=5"

	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return helpers.OperationFailure("Graph joinedTeams failed", err), err
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
