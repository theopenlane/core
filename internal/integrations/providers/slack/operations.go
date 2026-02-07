package slack

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	slackOperationHealth      types.OperationName = "health.default"
	slackOperationTeam        types.OperationName = "team.inspect"
	slackOperationMessagePost types.OperationName = "message.post"
)

type slackMessageOperationConfig struct {
	// Channel identifies the Slack channel or user to receive the message
	Channel string `json:"channel" jsonschema:"required,description=Slack channel ID or user ID to receive the message."`
	// Text is the message text when blocks are not supplied
	Text string `json:"text,omitempty" jsonschema:"description=Message text (required unless blocks are supplied)."`
	// Blocks carries optional Block Kit payloads
	Blocks []map[string]any `json:"blocks,omitempty" jsonschema:"description=Optional Slack Block Kit payload."`
	// Attachments carries optional legacy attachments
	Attachments []map[string]any `json:"attachments,omitempty" jsonschema:"description=Optional legacy attachments payload."`
	// ThreadTS identifies the thread timestamp for replies
	ThreadTS string `json:"thread_ts,omitempty" jsonschema:"description=Optional thread timestamp to reply within an existing thread."`
	// UnfurlLinks controls link unfurling in messages
	UnfurlLinks *bool `json:"unfurl_links,omitempty" jsonschema:"description=Whether to unfurl links in the message."`
	// UnfurlMedia controls media unfurling in messages
	UnfurlMedia *bool `json:"unfurl_media,omitempty" jsonschema:"description=Whether to unfurl media in the message."`
}

type slackMessageConfig struct {
	// Channel selects the Slack channel
	Channel types.TrimmedString `json:"channel"`
	// ChannelID is an alias for Channel
	ChannelID types.TrimmedString `json:"channel_id"`
	// Text is the message text input
	Text types.TrimmedString `json:"text"`
	// Message is an alternate message text input
	Message types.TrimmedString `json:"message"`
	// Body is an alternate message body input
	Body types.TrimmedString `json:"body"`
	// Blocks carries Block Kit payloads
	Blocks any `json:"blocks"`
	// Attachments carries legacy attachment payloads
	Attachments any `json:"attachments"`
	// ThreadTS identifies the thread timestamp for replies
	ThreadTS types.TrimmedString `json:"thread_ts"`
	// UnfurlLinks controls link unfurling
	UnfurlLinks *bool `json:"unfurl_links"`
	// UnfurlMedia controls media unfurling
	UnfurlMedia *bool `json:"unfurl_media"`
}

var slackMessageConfigSchema = operations.SchemaFrom[slackMessageOperationConfig]()

// slackOperations returns the Slack operations supported by this provider.
func slackOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(slackOperationHealth, "Call auth.test to ensure the Slack token is valid and scoped correctly.", ClientSlackAPI, runSlackHealthOperation),
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
	// OK indicates whether the request succeeded
	OK bool `json:"ok"`
	// URL is the workspace URL
	URL string `json:"url"`
	// Team is the workspace name
	Team string `json:"team"`
	// User is the user name associated with the token
	User string `json:"user"`
	// Error contains the error message when ok is false
	Error string `json:"error"`
}

// slackTeamInfoResponse represents the response from Slack team.info
type slackTeamInfoResponse struct {
	// OK indicates whether the request succeeded
	OK bool `json:"ok"`
	// Team holds the workspace details
	Team slackTeamInfo `json:"team"`
	// Error contains the error message when ok is false
	Error string `json:"error"`
}

// slackTeamInfo represents Slack workspace information
type slackTeamInfo struct {
	// ID is the workspace identifier
	ID string `json:"id"`
	// Name is the workspace name
	Name string `json:"name"`
	// Domain is the workspace domain
	Domain string `json:"domain"`
	// EmailDomain is the workspace email domain
	EmailDomain string `json:"email_domain"`
}

// runSlackHealthOperation verifies the Slack OAuth token via auth.test
func runSlackHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndOAuthToken(input, TypeSlack)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp slackAuthTestResponse
	if err := slackAPIGet(ctx, client, token, "auth.test", nil, &resp); err != nil {
		return operations.OperationFailure("Slack auth.test failed", err), err
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
	client, token, err := auth.ClientAndOAuthToken(input, TypeSlack)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp slackTeamInfoResponse
	if err := slackAPIGet(ctx, client, token, "team.info", nil, &resp); err != nil {
		return operations.OperationFailure("Slack team.info failed", err), err
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
	// OK indicates whether the message was posted
	OK bool `json:"ok"`
	// Channel is the channel identifier where the message was posted
	Channel string `json:"channel"`
	// TS is the message timestamp
	TS string `json:"ts"`
	// Error contains the error message when ok is false
	Error string `json:"error"`
}

// runSlackMessagePostOperation sends a message to a Slack channel or user
func runSlackMessagePostOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	_, token, err := auth.ClientAndOAuthToken(input, TypeSlack)
	if err != nil {
		return types.OperationResult{}, err
	}

	var cfg slackMessageConfig
	if err := operations.DecodeConfig(input.Config, &cfg); err != nil {
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
	if err := auth.HTTPPostJSON(ctx, nil, endpoint, token, nil, payload, &resp); err != nil {
		return operations.OperationFailure("Slack chat.postMessage failed", err), err
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
func slackAPIGet(ctx context.Context, client *auth.AuthenticatedClient, token, method string, params url.Values, out any) error {
	endpoint := "https://slack.com/api/" + method
	if params != nil {
		if query := params.Encode(); query != "" {
			endpoint += "?" + query
		}
	}

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, out); err != nil {
		if errors.Is(err, auth.ErrHTTPRequestFailed) {
			return ErrAPIRequest
		}
		return err
	}

	return nil
}
