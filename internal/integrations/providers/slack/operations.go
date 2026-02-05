package slack

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	slackOperationHealth      types.OperationName = "health.default"
	slackOperationTeam        types.OperationName = "team.inspect"
	slackOperationMessagePost types.OperationName = "message.post"
)

type slackMessageOperationConfig struct {
	Channel     string           `json:"channel" jsonschema:"required,description=Slack channel ID or user ID to receive the message."`
	Text        string           `json:"text,omitempty" jsonschema:"description=Message text (required unless blocks are supplied)."`
	Blocks      []map[string]any `json:"blocks,omitempty" jsonschema:"description=Optional Slack Block Kit payload."`
	Attachments []map[string]any `json:"attachments,omitempty" jsonschema:"description=Optional legacy attachments payload."`
	ThreadTS    string           `json:"thread_ts,omitempty" jsonschema:"description=Optional thread timestamp to reply within an existing thread."`
	UnfurlLinks *bool            `json:"unfurl_links,omitempty" jsonschema:"description=Whether to unfurl links in the message."`
	UnfurlMedia *bool            `json:"unfurl_media,omitempty" jsonschema:"description=Whether to unfurl media in the message."`
}

type slackMessageConfig struct {
	Channel     helpers.TrimmedString `mapstructure:"channel"`
	ChannelID   helpers.TrimmedString `mapstructure:"channel_id"`
	Text        helpers.TrimmedString `mapstructure:"text"`
	Message     helpers.TrimmedString `mapstructure:"message"`
	Body        helpers.TrimmedString `mapstructure:"body"`
	Blocks      any                   `mapstructure:"blocks"`
	Attachments any                   `mapstructure:"attachments"`
	ThreadTS    helpers.TrimmedString `mapstructure:"thread_ts"`
	UnfurlLinks *bool                 `mapstructure:"unfurl_links"`
	UnfurlMedia *bool                 `mapstructure:"unfurl_media"`
}

var slackMessageConfigSchema = helpers.SchemaFrom[slackMessageOperationConfig]()

// slackOperations returns the Slack operations supported by this provider.
func slackOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		helpers.HealthOperation(slackOperationHealth, "Call auth.test to ensure the Slack token is valid and scoped correctly.", ClientSlackAPI, runSlackHealthOperation),
		{
			Name:        slackOperationTeam,
			Kind:        types.OperationKindScanSettings,
			Description: "Collect workspace metadata via team.info for posture analysis.",
			Client:      ClientSlackAPI,
			Run:         runSlackTeamOperation,
		},
		{
			Name:         slackOperationMessagePost,
			Kind:         types.OperationKindNotify,
			Description:  "Send a Slack message via chat.postMessage.",
			Client:       ClientSlackAPI,
			Run:          runSlackMessagePostOperation,
			ConfigSchema: slackMessageConfigSchema,
		},
	}
}

// slackAuthTestResponse represents the response from Slack auth.test
type slackAuthTestResponse struct {
	OK    bool   `json:"ok"`
	URL   string `json:"url"`
	Team  string `json:"team"`
	User  string `json:"user"`
	Error string `json:"error"`
}

// slackTeamInfoResponse represents the response from Slack team.info
type slackTeamInfoResponse struct {
	OK    bool          `json:"ok"`
	Team  slackTeamInfo `json:"team"`
	Error string        `json:"error"`
}

// slackTeamInfo represents Slack workspace information
type slackTeamInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	EmailDomain string `json:"email_domain"`
}

// runSlackHealthOperation verifies the Slack OAuth token via auth.test
func runSlackHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, TypeSlack)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp slackAuthTestResponse
	if err := slackAPIGet(ctx, client, token, "auth.test", nil, &resp); err != nil {
		return helpers.OperationFailure("Slack auth.test failed", err), err
	}

	if !resp.OK {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack auth.test returned error",
			Details: map[string]any{"error": resp.Error},
		}, ErrSlackAPIError
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Slack token valid for workspace %s", resp.Team),
		Details: map[string]any{
			"team": resp.Team,
			"url":  resp.URL,
			"user": resp.User,
		},
	}, nil
}

// runSlackTeamOperation fetches workspace metadata for posture analysis
func runSlackTeamOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := helpers.ClientAndOAuthToken(input, TypeSlack)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp slackTeamInfoResponse
	if err := slackAPIGet(ctx, client, token, "team.info", nil, &resp); err != nil {
		return helpers.OperationFailure("Slack team.info failed", err), err
	}

	if !resp.OK {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack team.info returned error",
			Details: map[string]any{"error": resp.Error},
		}, ErrSlackAPIError
	}

	team := resp.Team
	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Workspace %s (%s) settings retrieved", team.Name, team.ID),
		Details: map[string]any{
			"teamId":      team.ID,
			"name":        team.Name,
			"domain":      team.Domain,
			"emailDomain": team.EmailDomain,
		},
	}, nil
}

type slackMessageResponse struct {
	OK      bool   `json:"ok"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
	Error   string `json:"error"`
}

// runSlackMessagePostOperation sends a message to a Slack channel or user
func runSlackMessagePostOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	_, token, err := helpers.ClientAndOAuthToken(input, TypeSlack)
	if err != nil {
		return types.OperationResult{}, err
	}

	var cfg slackMessageConfig
	if err := helpers.DecodeConfig(input.Config, &cfg); err != nil {
		return types.OperationResult{}, err
	}

	channel := string(cfg.Channel)
	if channel == "" {
		channel = string(cfg.ChannelID)
	}
	if channel == "" {
		return types.OperationResult{}, ErrSlackChannelMissing
	}

	payload := map[string]any{
		"channel": channel,
	}

	text := string(cfg.Text)
	if text == "" {
		text = string(cfg.Message)
	}
	if text == "" {
		text = string(cfg.Body)
	}
	if text != "" {
		payload["text"] = text
	}

	if cfg.Blocks != nil {
		payload["blocks"] = cfg.Blocks
	}
	if cfg.Attachments != nil {
		payload["attachments"] = cfg.Attachments
	}
	if cfg.ThreadTS != "" {
		payload["thread_ts"] = string(cfg.ThreadTS)
	}
	if cfg.UnfurlLinks != nil {
		payload["unfurl_links"] = cfg.UnfurlLinks
	}
	if cfg.UnfurlMedia != nil {
		payload["unfurl_media"] = cfg.UnfurlMedia
	}

	if _, ok := payload["text"]; !ok {
		if _, hasBlocks := payload["blocks"]; !hasBlocks {
			if _, hasAttachments := payload["attachments"]; !hasAttachments {
				return types.OperationResult{}, ErrSlackMessageEmpty
			}
		}
	}

	var resp slackMessageResponse
	endpoint := "https://slack.com/api/chat.postMessage"
	if err := helpers.HTTPPostJSON(ctx, nil, endpoint, token, nil, payload, &resp); err != nil {
		return helpers.OperationFailure("Slack chat.postMessage failed", err), err
	}

	if !resp.OK {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "Slack chat.postMessage returned error",
			Details: map[string]any{"error": resp.Error},
		}, ErrSlackAPIError
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Slack message sent to %s", resp.Channel),
		Details: map[string]any{
			"channel": resp.Channel,
			"ts":      resp.TS,
		},
	}, nil
}

// slackAPIGet performs a GET request to the Slack API and decodes the JSON response
func slackAPIGet(ctx context.Context, client *helpers.AuthenticatedClient, token, method string, params url.Values, out any) error {
	endpoint := "https://slack.com/api/" + method
	if params != nil {
		if query := params.Encode(); query != "" {
			endpoint += "?" + query
		}
	}

	if err := helpers.GetJSONWithClient(ctx, client, endpoint, token, nil, out); err != nil {
		if errors.Is(err, helpers.ErrHTTPRequestFailed) {
			return fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}
		return err
	}

	return nil
}
