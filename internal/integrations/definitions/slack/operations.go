package slack

import (
	"context"
	"encoding/json"
	"fmt"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// HealthCheck holds the result of a Slack health check
type HealthCheck struct {
	// Team is the Slack team name
	Team string `json:"team"`
	// URL is the Slack team URL
	URL string `json:"url"`
	// User is the authenticated Slack user
	User string `json:"user"`
}

// TeamInspect collects Slack workspace metadata via team.info
type TeamInspect struct {
	// TeamID is the Slack team identifier
	TeamID string `json:"teamId"`
	// Name is the Slack team display name
	Name string `json:"name"`
	// Domain is the Slack team domain
	Domain string `json:"domain"`
	// EmailDomain is the team's verified email domain
	EmailDomain string `json:"emailDomain"`
}

// MessageSend sends a Slack message via chat.postMessage
type MessageSend struct {
	// Channel is the channel the message was posted to
	Channel string `json:"channel"`
	// TS is the message timestamp
	TS string `json:"ts"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, c)
	}
}

// Run executes the Slack auth.test health check
func (HealthCheck) Run(ctx context.Context, c *slackgo.Client) (json.RawMessage, error) {
	resp, err := c.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("slack: auth.test failed: %w", err)
	}

	return jsonx.ToRawMessage(HealthCheck{
		Team: resp.Team,
		URL:  resp.URL,
		User: resp.User,
	})
}

// Handle adapts team inspect to the generic operation registration boundary
func (t TeamInspect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return t.Run(ctx, c)
	}
}

// Run collects Slack workspace metadata via team.info
func (TeamInspect) Run(ctx context.Context, c *slackgo.Client) (json.RawMessage, error) {
	team, err := c.GetTeamInfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("slack: team.info failed: %w", err)
	}

	return jsonx.ToRawMessage(TeamInspect{
		TeamID:      team.ID,
		Name:        team.Name,
		Domain:      team.Domain,
		EmailDomain: team.EmailDomain,
	})
}

// Handle adapts message send to the generic operation registration boundary
func (m MessageSend) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var cfg MessageOperationInput
		if err := jsonx.UnmarshalIfPresent(request.Config, &cfg); err != nil {
			return nil, err
		}

		return m.Run(ctx, c, cfg)
	}
}

// Run sends a Slack message via chat.postMessage
func (MessageSend) Run(ctx context.Context, c *slackgo.Client, cfg MessageOperationInput) (json.RawMessage, error) {
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

	return jsonx.ToRawMessage(MessageSend{
		Channel: respChannel,
		TS:      ts,
	})
}
