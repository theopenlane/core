package microsoftteams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const teamsGraphBaseURL = "https://graph.microsoft.com/v1.0/"

const (
	teamsHealthOp      types.OperationName = types.OperationHealthDefault
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

var teamsMessageConfigSchema = operations.SchemaFrom[teamsMessageOperationConfig]()

// teamsOperations returns the Microsoft Teams operations supported by this provider
func teamsOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(teamsHealthOp, "Call Graph /me to verify Teams access.", ClientMicrosoftTeamsAPI,
			operations.HealthCheckRunner(auth.OAuthTokenFromPayload, "https://graph.microsoft.com/v1.0/me", "Graph /me failed",
				func(profile teamsProfileResponse) (string, any) {
					return fmt.Sprintf("Graph token valid for %s", profile.DisplayName), teamsHealthDetails{
						ID:   profile.ID,
						Mail: profile.Mail,
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

type teamsHealthDetails struct {
	ID   string `json:"id"`
	Mail string `json:"mail"`
}

type teamsSampleEntry struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type teamsSampleDetails struct {
	Teams []teamsSampleEntry `json:"teams"`
}

type teamsMessageSendDetails struct {
	TeamID    string `json:"teamId"`
	ChannelID string `json:"channelId"`
	MessageID string `json:"messageId"`
}

// runTeamsSample collects a sample of joined Teams for the authenticated user
func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := auth.ResolveAuthenticatedClient(input, auth.OAuthTokenFromPayload, teamsGraphBaseURL, nil)
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

	if err := client.GetJSON(ctx, "me/joinedTeams?$top=5", &resp); err != nil {
		return operations.OperationFailure("Graph joinedTeams failed", err, nil)
	}

	samples := make([]teamsSampleEntry, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, teamsSampleEntry{
			ID:          team.ID,
			DisplayName: team.DisplayName,
		})
	}

	return operations.OperationSuccess(fmt.Sprintf("Retrieved %d joined teams", len(samples)), teamsSampleDetails{Teams: samples}), nil
}

// runTeamsMessageSendOperation posts a message to a Teams channel
func runTeamsMessageSendOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := auth.ResolveAuthenticatedClient(input, auth.OAuthTokenFromPayload, teamsGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	var cfg teamsMessageOperationConfig
	if len(input.Config) > 0 {
		if err := json.Unmarshal(input.Config, &cfg); err != nil {
			return types.OperationResult{}, err
		}
	}

	teamID := cfg.TeamID
	channelID := cfg.ChannelID
	if teamID == "" || channelID == "" {
		return types.OperationResult{}, ErrTeamsChannelMissing
	}

	body := cfg.Body
	if body == "" {
		return types.OperationResult{}, ErrTeamsMessageEmpty
	}

	contentType := cfg.BodyFormat
	if contentType == "" {
		contentType = "text"
	}
	if contentType != "text" && contentType != "html" {
		return types.OperationResult{}, ErrTeamsMessageFormatInvalid
	}

	payload := map[string]any{
		"body": map[string]any{
			"contentType": contentType,
			"content":     body,
		},
	}

	if cfg.Subject != "" {
		payload["subject"] = cfg.Subject
	}

	path := fmt.Sprintf("teams/%s/channels/%s/messages", url.PathEscape(teamID), url.PathEscape(channelID))
	var resp struct {
		ID string `json:"id"`
	}
	if err := client.PostJSON(ctx, path, payload, &resp); err != nil {
		return operations.OperationFailure("Graph channel message failed", err, nil)
	}

	return operations.OperationSuccess(fmt.Sprintf("Teams message sent to %s", channelID), teamsMessageSendDetails{
		TeamID:    teamID,
		ChannelID: channelID,
		MessageID: resp.ID,
	}), nil
}
