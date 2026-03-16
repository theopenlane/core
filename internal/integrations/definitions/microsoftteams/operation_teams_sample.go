package microsoftteams

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// SampleEntry holds a single Teams team entry
type SampleEntry struct {
	// ID is the Teams team identifier
	ID string `json:"id"`
	// DisplayName is the team display name
	DisplayName string `json:"displayName"`
}

// TeamsSample collects a sample of joined Microsoft Teams
type TeamsSample struct {
	// Teams is the collected team sample
	Teams []SampleEntry `json:"teams"`
}

// Handle adapts teams sample to the generic operation registration boundary
func (t TeamsSample) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return t.Run(ctx, c)
	}
}

// Run collects a sample of joined Microsoft Teams
func (TeamsSample) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	var resp struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	if err := c.GetJSON(ctx, "me/joinedTeams?$top=5", &resp); err != nil {
		return nil, ErrJoinedTeamsLookupFailed
	}

	samples := make([]SampleEntry, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, SampleEntry{
			ID:          team.ID,
			DisplayName: team.DisplayName,
		})
	}

	return providerkit.EncodeResult(TeamsSample{Teams: samples}, ErrResultEncode)
}
