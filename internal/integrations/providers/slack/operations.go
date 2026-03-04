package slack

import (
	"context"
	"fmt"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	slackOperationHealth      types.OperationName = "health.default"
	slackOperationTeam        types.OperationName = "team.inspect"
	slackOperationMessageSend types.OperationName = "message.send"
)

type slackMessageOperationConfig struct {
	// Channel identifies the Slack channel or user to receive the message
	Channel string `json:"channel" jsonschema:"required,description=Slack channel ID or user ID to receive the message."`
	// Text is the message text when blocks are not supplied
	Text string `json:"text,omitempty" jsonschema:"description=Message text (required unless blocks are supplied)."`
	// Blocks carries optional Block Kit payloads
	Blocks []map[string]any `json:"blocks,omitempty" jsonschema:"description=Optional Slack Block Kit payload."`
	// Attachments carries optional attachments payloads
	Attachments []map[string]any `json:"attachments,omitempty" jsonschema:"description=Optional attachments payload."`
	// ThreadTS identifies the thread timestamp for replies
	ThreadTS string `json:"thread_ts,omitempty" jsonschema:"description=Optional thread timestamp to reply within an existing thread."`
	// UnfurlLinks controls link unfurling in messages
	UnfurlLinks *bool `json:"unfurl_links,omitempty" jsonschema:"description=Whether to unfurl links in the message."`
	// UnfurlMedia controls media unfurling in messages
	UnfurlMedia *bool `json:"unfurl_media,omitempty" jsonschema:"description=Whether to unfurl media in the message."`
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
			Name:         slackOperationMessageSend,
			Kind:         types.OperationKindNotify,
			Description:  "Send a Slack message via chat.postMessage.",
			Client:       ClientSlackAPI,
			Run:          runSlackMessagePostOperation,
			ConfigSchema: slackMessageConfigSchema,
		},
	}
}

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

// resolveSlackClient returns a pooled Slack client or builds one from the credential payload.
func resolveSlackClient(input types.OperationInput) (*slackgo.Client, error) {
	if c, ok := types.ClientInstanceAs[*slackgo.Client](input.Client); ok {
		return c, nil
	}

	token, err := auth.OAuthTokenFromPayload(input.Credential)
	if err != nil {
		return nil, err
	}

	return slackgo.New(token), nil
}

// runSlackHealthOperation verifies the Slack OAuth token via auth.test
func runSlackHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveSlackClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := client.AuthTestContext(ctx)
	if err != nil {
		return operations.OperationFailure("Slack auth.test failed", err, nil)
	}

	return operations.OperationSuccess(fmt.Sprintf("Slack token valid for workspace %s", resp.Team), slackHealthDetails{
		Team: resp.Team,
		URL:  resp.URL,
		User: resp.User,
	}), nil
}

// runSlackTeamOperation fetches workspace metadata for posture analysis
func runSlackTeamOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveSlackClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	team, err := client.GetTeamInfoContext(ctx)
	if err != nil {
		return operations.OperationFailure("Slack team.info failed", err, nil)
	}

	return operations.OperationSuccess(fmt.Sprintf("Workspace %s (%s) settings retrieved", team.Name, team.ID), slackTeamDetails{
		TeamID:      team.ID,
		Name:        team.Name,
		Domain:      team.Domain,
		EmailDomain: team.EmailDomain,
	}), nil
}

// runSlackMessagePostOperation sends a message to a Slack channel or user
func runSlackMessagePostOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveSlackClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	cfg, err := operations.Decode[slackMessageOperationConfig](input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	channel := cfg.Channel
	if channel == "" {
		return types.OperationResult{}, ErrSlackChannelMissing
	}

	hasText := cfg.Text != ""
	hasBlocks := len(cfg.Blocks) > 0
	hasAttachments := len(cfg.Attachments) > 0

	if !hasText && !hasBlocks && !hasAttachments {
		return types.OperationResult{}, ErrSlackMessageEmpty
	}

	opts := []slackgo.MsgOption{
		slackgo.MsgOptionAsUser(true),
	}

	if hasText {
		opts = append(opts, slackgo.MsgOptionText(cfg.Text, false))
	}

	if cfg.ThreadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(cfg.ThreadTS))
	}

	if cfg.UnfurlLinks != nil && !*cfg.UnfurlLinks {
		opts = append(opts, slackgo.MsgOptionDisableLinkUnfurl())
	}

	respChannel, ts, err := client.PostMessageContext(ctx, channel, opts...)
	if err != nil {
		return operations.OperationFailure("Slack chat.postMessage failed", err, nil)
	}

	return operations.OperationSuccess(fmt.Sprintf("Slack message sent to %s", respChannel), slackMessageDetails{
		Channel: respChannel,
		TS:      ts,
	}), nil
}
