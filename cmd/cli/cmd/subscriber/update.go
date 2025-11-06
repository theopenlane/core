//go:build cli

package subscribers

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildSubscriberUpdateInput() (string, openlaneclient.UpdateSubscriberInput, error) {
	email := cmd.Config.String("email")
	if email == "" {
		return "", openlaneclient.UpdateSubscriberInput{}, cmd.NewRequiredFieldMissingError("email")
	}

	phone := cmd.Config.String("phone-number")
	input := openlaneclient.UpdateSubscriberInput{PhoneNumber: &phone}

	return email, input, nil
}

func updateSubscriber(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateSubscriber, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	email, input, err := buildSubscriberUpdateInput()
	if err != nil {
		return nil, err
	}

	return client.UpdateSubscriber(ctx, email, input)
}
