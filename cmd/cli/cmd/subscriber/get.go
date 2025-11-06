//go:build cli

package subscribers

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func getSubscribers(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	email := cmd.Config.String("email")
	if email != "" {
		return client.GetSubscriberByEmail(ctx, email)
	}

	active := cmd.Config.Bool("active")
	where := openlaneclient.SubscriberWhereInput{}
	if cmd.Config.Exists("active") {
		where.Active = &active
	}

	return client.GetSubscribers(ctx, nil, nil, nil, nil, &where, nil)
}
