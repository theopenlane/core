package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Client builds Slack Web API clients for one installation
type Client struct{}

// Build constructs the Slack Web API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred slackCred
	if err := jsonx.UnmarshalIfPresent(req.Credential.Data, &cred); err != nil {
		return nil, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	return slackgo.New(cred.AccessToken), nil
}
