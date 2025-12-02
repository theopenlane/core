package microsoftteams

import (
	"context"
	"fmt"

	"github.com/theopenlane/shared/integrations/providers/helpers"
	"github.com/theopenlane/shared/integrations/types"
)

const (
	teamsHealthOp   types.OperationName = "health.default"
	teamsChannelsOp types.OperationName = "teams.sample"
)

func teamsOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        teamsHealthOp,
			Kind:        types.OperationKindHealth,
			Description: "Call Graph /me to verify Teams access.",
			Run:         runTeamsHealth,
		},
		{
			Name:        teamsChannelsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of joined teams for the user context.",
			Run:         runTeamsSample,
		},
	}
}

func runTeamsHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeMicrosoftTeams))
	if err != nil {
		return types.OperationResult{}, err
	}

	var profile struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
	}

	if err := helpers.HTTPGetJSON(ctx, nil, "https://graph.microsoft.com/v1.0/me", token, nil, &profile); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Graph /me failed",
			Details: map[string]any{"error": err.Error()},
		}, err
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Graph token valid for %s", profile.DisplayName),
		Details: map[string]any{"id": profile.ID, "mail": profile.Mail},
	}, nil
}

func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeMicrosoftTeams))
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
	if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp); err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Graph joinedTeams failed",
			Details: map[string]any{"error": err.Error()},
		}, err
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
