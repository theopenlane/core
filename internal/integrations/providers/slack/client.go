package slack

import (
	"context"
	"encoding/json"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientSlackAPI identifies the Slack HTTP API client.
	ClientSlackAPI types.ClientName = "api"
)

// slackClientDescriptors returns the client descriptors published by Slack.
func slackClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeSlack, ClientSlackAPI, "Slack Web API client", buildSlackClient)
}

// buildSlackClient constructs a Slack SDK client from credential payload.
func buildSlackClient(_ context.Context, payload models.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := auth.OAuthTokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(slackgo.New(token)), nil
}
