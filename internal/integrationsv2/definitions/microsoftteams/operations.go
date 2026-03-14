package microsoftteams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const teamsGraphBaseURL = "https://graph.microsoft.com/v1.0/"

type teamsProfileResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Mail        string `json:"mail"`
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

type teamsChannelMessageBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type teamsChannelMessageRequest struct {
	Body    teamsChannelMessageBody `json:"body"`
	Subject string                  `json:"subject,omitempty"`
}

// buildGraphClient builds the Microsoft Graph API client for one installation
func buildGraphClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	return providerkit.NewAuthenticatedClient(teamsGraphBaseURL, token, nil), nil
}

// runHealthOperation calls /me to verify Teams access
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	var profile teamsProfileResponse
	if err := c.GetJSON(ctx, "me", &profile); err != nil {
		return nil, fmt.Errorf("microsoftteams: graph /me failed: %w", err)
	}

	return jsonx.ToRawMessage(teamsHealthDetails{
		ID:   profile.ID,
		Mail: profile.Mail,
	})
}

// runTeamsSampleOperation collects a sample of joined Teams for the user
func runTeamsSampleOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	var resp struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	if err := c.GetJSON(ctx, "me/joinedTeams?$top=5", &resp); err != nil {
		return nil, fmt.Errorf("microsoftteams: graph joinedTeams failed: %w", err)
	}

	samples := make([]teamsSampleEntry, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, teamsSampleEntry{
			ID:          team.ID,
			DisplayName: team.DisplayName,
		})
	}

	return jsonx.ToRawMessage(teamsSampleDetails{Teams: samples})
}

// runMessageSendOperation posts a message to a Teams channel
func runMessageSendOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	c, ok := client.(*providerkit.AuthenticatedClient)
	if !ok {
		return nil, ErrClientType
	}

	var cfg messageOperationInput
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	if cfg.TeamID == "" || cfg.ChannelID == "" {
		return nil, ErrChannelMissing
	}

	if cfg.Body == "" {
		return nil, ErrMessageEmpty
	}

	contentType := cfg.BodyFormat
	if contentType == "" {
		contentType = "text"
	}

	switch contentType {
	case "text", "html":
	default:
		return nil, ErrBodyFormatInvalid
	}

	payload := teamsChannelMessageRequest{
		Body: teamsChannelMessageBody{
			ContentType: contentType,
			Content:     cfg.Body,
		},
		Subject: cfg.Subject,
	}

	path := fmt.Sprintf("teams/%s/channels/%s/messages", url.PathEscape(cfg.TeamID), url.PathEscape(cfg.ChannelID))

	var resp struct {
		ID string `json:"id"`
	}

	if err := c.PostJSON(ctx, path, payload, &resp); err != nil {
		return nil, fmt.Errorf("microsoftteams: graph channel message failed: %w", err)
	}

	return jsonx.ToRawMessage(teamsMessageSendDetails{
		TeamID:    cfg.TeamID,
		ChannelID: cfg.ChannelID,
		MessageID: resp.ID,
	})
}
