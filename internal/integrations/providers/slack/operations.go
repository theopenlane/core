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

// slackOperations returns the Slack operations supported by this provider.
func slackOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        slackOperationHealth,
			Kind:        types.OperationKindHealth,
			Description: "Call auth.test to ensure the Slack token is valid and scoped correctly.",
			Client:      ClientSlackAPI,
			Run:         runSlackHealthOperation,
		},
		{
			Name:        slackOperationTeam,
			Kind:        types.OperationKindScanSettings,
			Description: "Collect workspace metadata via team.info for posture analysis.",
			Client:      ClientSlackAPI,
			Run:         runSlackTeamOperation,
		},
		{
			Name:        slackOperationMessagePost,
			Kind:        types.OperationKindNotify,
			Description: "Send a Slack message via chat.postMessage.",
			Client:      ClientSlackAPI,
			Run:         runSlackMessagePostOperation,
			ConfigSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"channel": map[string]any{
						"type":        "string",
						"description": "Slack channel ID or user ID to receive the message.",
					},
					"text": map[string]any{
						"type":        "string",
						"description": "Message text (required unless blocks are supplied).",
					},
					"blocks": map[string]any{
						"type":        "array",
						"description": "Optional Slack Block Kit payload.",
					},
					"attachments": map[string]any{
						"type":        "array",
						"description": "Optional legacy attachments payload.",
					},
					"thread_ts": map[string]any{
						"type":        "string",
						"description": "Optional thread timestamp to reply within an existing thread.",
					},
				},
				"required": []string{"channel"},
			},
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

	channel := helpers.FirstStringValue(input.Config, "channel", "channel_id", "channelId")
	if channel == "" {
		return types.OperationResult{}, ErrSlackChannelMissing
	}

	payload := map[string]any{
		"channel": channel,
	}

	if text := helpers.FirstStringValue(input.Config, "text", "message", "body"); text != "" {
		payload["text"] = text
	}
	if blocks, ok := input.Config["blocks"]; ok {
		payload["blocks"] = blocks
	}
	if attachments, ok := input.Config["attachments"]; ok {
		payload["attachments"] = attachments
	}
	if threadTS, ok := input.Config["thread_ts"]; ok {
		payload["thread_ts"] = threadTS
	}
	if unfurlLinks, ok := input.Config["unfurl_links"]; ok {
		payload["unfurl_links"] = unfurlLinks
	}
	if unfurlMedia, ok := input.Config["unfurl_media"]; ok {
		payload["unfurl_media"] = unfurlMedia
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

// oauthTokenFromPayload extracts the OAuth access token from the credential payload
func oauthTokenFromPayload(payload types.CredentialPayload) (string, error) {
	tokenOpt := payload.OAuthTokenOption()
	if !tokenOpt.IsPresent() {
		return "", ErrOAuthTokenMissing
	}

	token := tokenOpt.MustGet()
	if token == nil || token.AccessToken == "" {
		return "", ErrAccessTokenEmpty
	}

	return token.AccessToken, nil
}
