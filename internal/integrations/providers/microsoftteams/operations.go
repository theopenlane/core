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

// teamsOperations returns the Microsoft Teams operations supported by this provider.
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

func runTeamsHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
	token, err := helpers.OAuthTokenFromPayload(input.Credential, string(TypeMicrosoftTeams))
	if err != nil {
		return types.OperationResult{}, err
	}

	var profile struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
	}

	endpoint := "https://graph.microsoft.com/v1.0/me"
	if client != nil {
		if err := client.GetJSON(ctx, endpoint, &profile); err != nil {
			return types.OperationResult{
				Status:  types.OperationStatusFailed,
				Summary: "Graph /me failed",
				Details: map[string]any{"error": err.Error()},
			}, err
		}
	} else if err := helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &profile); err != nil {
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

// runTeamsSample collects a sample of joined Teams for the authenticated user.
func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client := helpers.AuthenticatedClientFromAny(input.Client)
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

	if client != nil {
		err = client.GetJSON(ctx, endpoint, &resp)
	} else {
		err = helpers.HTTPGetJSON(ctx, nil, endpoint, token, nil, &resp)
	}
	if err != nil {
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
