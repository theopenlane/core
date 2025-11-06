//go:build cli

package usersetting

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// getUserSettings retrieves user settings using the legacy semantics.
func getUserSettings(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	// filter options
	id := cmd.Config.String("id")

	// if setting ID is not provided, get settings which will automatically filter by user id
	if id != "" {
		return client.GetUserSettingByID(ctx, id)
	}

	return client.GetAllUserSettings(ctx)
}
