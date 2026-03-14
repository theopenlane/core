package slack

import (
	"context"
	"encoding/json"
	"fmt"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

type slackHealthDetails struct {
	Team string `json:"team"`
	URL  string `json:"url"`
	User string `json:"user"`
}

type slackTeamDetails struct {
	TeamID      string `json:"teamId"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	EmailDomain string `json:"emailDomain"`
}

type slackMessageDetails struct {
	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

// buildSlackClient builds the Slack Web API client for one installation
func buildSlackClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	return slackgo.New(token), nil
}

// runHealthOperation calls auth.test to verify the Slack token
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*slackgo.Client)
	if !ok {
		return nil, ErrClientType
	}

	resp, err := c.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("slack: auth.test failed: %w", err)
	}

	return jsonx.ToRawMessage(slackHealthDetails{
		Team: resp.Team,
		URL:  resp.URL,
		User: resp.User,
	})
}

// runTeamInspectOperation calls team.info to collect workspace metadata
func runTeamInspectOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*slackgo.Client)
	if !ok {
		return nil, ErrClientType
	}

	team, err := c.GetTeamInfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("slack: team.info failed: %w", err)
	}

	return jsonx.ToRawMessage(slackTeamDetails{
		TeamID:      team.ID,
		Name:        team.Name,
		Domain:      team.Domain,
		EmailDomain: team.EmailDomain,
	})
}

// runMessageSendOperation sends a Slack message via chat.postMessage
func runMessageSendOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*slackgo.Client)
	if !ok {
		return nil, ErrClientType
	}

	var cfg messageOperationInput
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	if cfg.Channel == "" {
		return nil, ErrChannelMissing
	}

	hasText := cfg.Text != ""
	hasBlocks := len(cfg.Blocks) > 0
	if !hasText && !hasBlocks {
		return nil, ErrMessageEmpty
	}

	opts := []slackgo.MsgOption{slackgo.MsgOptionAsUser(true)}

	if hasText {
		opts = append(opts, slackgo.MsgOptionText(cfg.Text, false))
	}

	if cfg.ThreadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(cfg.ThreadTS))
	}

	if cfg.UnfurlLinks != nil && !*cfg.UnfurlLinks {
		opts = append(opts, slackgo.MsgOptionDisableLinkUnfurl())
	}

	respChannel, ts, err := c.PostMessageContext(ctx, cfg.Channel, opts...)
	if err != nil {
		return nil, fmt.Errorf("slack: chat.postMessage failed: %w", err)
	}

	return jsonx.ToRawMessage(slackMessageDetails{
		Channel: respChannel,
		TS:      ts,
	})
}
