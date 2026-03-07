package microsoftteams

import (
	"context"
	"fmt"
	"net/url"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	teamsHealthOp      types.OperationName = "health.default"
	teamsChannelsOp    types.OperationName = "teams.sample"
	teamsMessageSendOp types.OperationName = "message.send"
)

type teamsMessageOperationConfig struct {
	// TeamID identifies the Team to post into
	TeamID string `json:"team_id" jsonschema:"required,description=Microsoft Teams team ID to receive the message."`
	// ChannelID identifies the channel within the team
	ChannelID string `json:"channel_id" jsonschema:"required,description=Microsoft Teams channel ID to receive the message."`
	// Body is the message body content
	Body string `json:"body" jsonschema:"required,description=Message body content."`
	// BodyFormat is the message format (text or html)
	BodyFormat string `json:"body_format,omitempty" jsonschema:"description=Message body format (text or html)."`
	// Subject is an optional message subject
	Subject string `json:"subject,omitempty" jsonschema:"description=Optional message subject."`
}

type teamsMessageConfig struct {
	// TeamID identifies the Team to post into
	TeamID types.TrimmedString `json:"team_id"`
	// Team is an alias for TeamID
	Team types.TrimmedString `json:"team"`
	// ChannelID identifies the channel within the team
	ChannelID types.TrimmedString `json:"channel_id"`
	// Channel is an alias for ChannelID
	Channel types.TrimmedString `json:"channel"`
	// Body is the message body content
	Body types.TrimmedString `json:"body"`
	// Text is an alias for Body
	Text types.TrimmedString `json:"text"`
	// Message is an alias for Body
	Message types.TrimmedString `json:"message"`
	// Subject is an optional message subject
	Subject types.TrimmedString `json:"subject"`
	// BodyFormat is the message format (text or html)
	BodyFormat types.LowerString `json:"body_format"`
}

var teamsMessageConfigSchema = operations.SchemaFrom[teamsMessageOperationConfig]()

// teamsOperations returns the Microsoft Teams operations supported by this provider
func teamsOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(teamsHealthOp, "Call Graph /me to verify Teams access.", ClientMicrosoftTeamsAPI,
			operations.HealthCheckRunner(operations.TokenTypeOAuth, "https://graph.microsoft.com/v1.0/me", "Graph /me failed",
				func(profile teamsProfileResponse) (string, map[string]any) {
					return fmt.Sprintf("Graph token valid for %s", profile.DisplayName), map[string]any{
						"id":   profile.ID,
						"mail": profile.Mail,
					}
				})),
		{
			Name:        teamsChannelsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of joined teams for the user context.",
			Client:      ClientMicrosoftTeamsAPI,
			Run:         runTeamsSample,
		},
		{
			Name:         teamsMessageSendOp,
			Kind:         types.OperationKindNotify,
			Description:  "Send a Teams channel message via Microsoft Graph.",
			Client:       ClientMicrosoftTeamsAPI,
			Run:          runTeamsMessageSendOperation,
			ConfigSchema: teamsMessageConfigSchema,
		},
	}
}

type teamsProfileResponse struct {
	// ID is the user identifier
	ID string `json:"id"`
	// DisplayName is the user display name
	DisplayName string `json:"displayName"`
	// Mail is the primary email address
	Mail string `json:"mail"`
}

// runTeamsSample collects a sample of joined Teams for the authenticated user
func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndToken(input, auth.OAuthTokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		// Value lists the joined teams
		Value []struct {
			// ID is the team identifier
			ID string `json:"id"`
			// DisplayName is the team display name
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	endpoint := "https://graph.microsoft.com/v1.0/me/joinedTeams?$top=5"

	if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
		return operations.OperationFailure("Graph joinedTeams failed", err, nil)
	}

	samples := make([]map[string]any, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, map[string]any{
			"id":          team.ID,
			"displayName": team.DisplayName,
		})
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Retrieved %d joined teams", len(samples)),
		Details: map[string]any{"teams": samples},
	}, nil
}

// runTeamsMessageSendOperation posts a message to a Teams channel
func runTeamsMessageSendOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	_, token, err := auth.ClientAndToken(input, auth.OAuthTokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	cfg, err := operations.Decode[teamsMessageConfig](input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	teamID := lo.CoalesceOrEmpty(cfg.TeamID, cfg.Team).String()
	channelID := lo.CoalesceOrEmpty(cfg.ChannelID, cfg.Channel).String()
	if teamID == "" || channelID == "" {
		return types.OperationResult{}, ErrTeamsChannelMissing
	}

	body := lo.CoalesceOrEmpty(cfg.Body, cfg.Text, cfg.Message).String()
	if body == "" {
		return types.OperationResult{}, ErrTeamsMessageEmpty
	}

	contentType := lo.CoalesceOrEmpty(cfg.BodyFormat, "text").String()
	if contentType != "text" && contentType != "html" {
		return types.OperationResult{}, ErrTeamsMessageFormatInvalid
	}

	payload := map[string]any{
		"body": map[string]any{
			"contentType": contentType,
			"content":     body,
		},
	}

	subject := cfg.Subject.String()
	if subject != "" {
		payload["subject"] = subject
	}

	endpoint := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels/%s/messages", url.PathEscape(teamID), url.PathEscape(channelID))
	var resp struct {
		ID string `json:"id"`
	}
	if err := auth.HTTPPostJSON(ctx, nil, endpoint, token, nil, payload, &resp); err != nil {
		return operations.OperationFailure("Graph channel message failed", err, nil)
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Teams message sent to %s", channelID),
		Details: map[string]any{
			"teamId":    teamID,
			"channelId": channelID,
			"messageId": resp.ID,
		},
	}, nil
}
