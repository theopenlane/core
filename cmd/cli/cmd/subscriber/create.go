//go:build cli

package subscribers

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildSubscriberInputs() ([]*openlaneclient.CreateSubscriberInput, error) {
	emails := cmd.Config.Strings("emails")
	if len(emails) == 0 {
		return nil, cmd.NewRequiredFieldMissingError("emails")
	}

	tags := cmd.Config.Strings("tags")

	inputs := make([]*openlaneclient.CreateSubscriberInput, 0, len(emails))
	for _, email := range emails {
		if email == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}

		in := &openlaneclient.CreateSubscriberInput{
			Email: email,
		}
		if len(tags) > 0 {
			in.Tags = tags
		}

		inputs = append(inputs, in)
	}

	return inputs, nil
}

func createSubscribers(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.CreateBulkSubscriber, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	inputs, err := buildSubscriberInputs()
	if err != nil {
		return nil, err
	}

	return client.CreateBulkSubscriber(ctx, inputs)
}
