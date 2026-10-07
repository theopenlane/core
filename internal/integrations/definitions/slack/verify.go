package slack

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify probes the workspace with auth.test and returns the installation identity
func verify[T any](ctx context.Context, _ types.ConnectionRequest[T], c *SlackClient) (InstallationMetadata, error) {
	resp, err := c.API.AuthTestContext(ctx)
	if err != nil {
		return InstallationMetadata{}, ErrAuthTestFailed
	}

	if resp.TeamID == "" && resp.Team == "" {
		return InstallationMetadata{}, ErrTeamIDMissing
	}

	return InstallationMetadata{TeamID: resp.TeamID, TeamName: resp.Team}, nil
}
