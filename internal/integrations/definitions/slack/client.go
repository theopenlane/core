package slack

import (
	"context"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Client builds Slack Web API clients for one installation
type Client struct{}

// Build constructs the Slack Web API client for one installation
func (Client) Build(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token, err := resolveAccessToken(req.Credentials)
	if err != nil {
		return nil, err
	}

	return slackgo.New(token), nil
}

// resolveAccessToken returns a usable Slack API token from whichever credential slot is bound
func resolveAccessToken(bindings types.CredentialBindings) (string, error) {
	if oauthCred, ok, err := slackCredential.Resolve(bindings); err != nil {
		return "", ErrCredentialDecode
	} else if ok {
		if oauthCred.AccessToken == "" {
			return "", ErrOAuthTokenMissing
		}

		return oauthCred.AccessToken, nil
	}

	if botCred, ok, err := slackBotTokenCredential.Resolve(bindings); err != nil {
		return "", ErrCredentialDecode
	} else if ok {
		if botCred.BotToken == "" {
			return "", ErrBotTokenMissing
		}

		return botCred.BotToken, nil
	}

	return "", ErrNoCredentialResolved
}
