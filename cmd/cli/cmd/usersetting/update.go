//go:build cli

package usersetting

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// updateValidation validates the input flags provided by the user.
func updateValidation() (id string, input openlaneclient.UpdateUserSettingInput, err error) {
	id = cmd.Config.String("id")

	// set default org if provided
	defaultOrg := cmd.Config.String("default-org")
	if defaultOrg != "" {
		input.DefaultOrgID = &defaultOrg
	}

	// set silenced at time if silence flag is set
	if cmd.Config.Bool("silence-notifications") {
		now := time.Now().UTC()
		input.SilencedAt = &now
	} else {
		input.SilencedAt = nil
	}

	// explicitly set 2fa if provided to avoid wiping setting unintentionally
	if cmd.Config.Bool("enable-2fa") {
		input.IsTfaEnabled = lo.ToPtr(true)
	} else if cmd.Config.Bool("disable-2fa") {
		input.IsTfaEnabled = lo.ToPtr(false)
	}

	// add tags to the input if provided
	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	// add status to the input if provided
	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToUserStatus(status)
	}

	return id, input, nil
}

// updateUserSetting executes the update mutation with the legacy semantics.
func updateUserSetting(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateUserSetting, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id, input, err := updateValidation()
	if err != nil {
		return nil, err
	}

	if id == "" {
		// get the user settings id
		settings, err := client.GetAllUserSettings(ctx)
		if err != nil {
			return nil, err
		}

		// this should never happen, but just in case
		if len(settings.GetUserSettings().Edges) == 0 {
			return nil, cmd.ErrNotFound
		}

		id = settings.GetUserSettings().Edges[0].Node.ID
	}

	// update the user settings
	return client.UpdateUserSetting(ctx, id, input)
}
