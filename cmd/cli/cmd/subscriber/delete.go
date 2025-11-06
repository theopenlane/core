//go:build cli

package subscribers

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildSubscriberDeleteInput() (string, *string, error) {
	email := cmd.Config.String("email")
	if email == "" {
		return "", nil, cmd.NewRequiredFieldMissingError("email")
	}

	orgID := cmd.Config.String("organization-id")
	if orgID == "" {
		return email, nil, nil
	}

	return email, &orgID, nil
}

func deleteSubscriber(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteSubscriber, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	email, orgID, err := buildSubscriberDeleteInput()
	if err != nil {
		return nil, err
	}

	return client.DeleteSubscriber(ctx, email, orgID)
}
