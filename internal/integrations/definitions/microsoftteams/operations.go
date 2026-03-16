package microsoftteams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// ProfileResponse is the raw Microsoft Graph /me response shape
type ProfileResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Mail        string `json:"mail"`
}

// HealthCheck holds the result of a Microsoft Teams health check
type HealthCheck struct {
	// ID is the Microsoft Graph user identifier
	ID string `json:"id"`
	// Mail is the user's email address
	Mail string `json:"mail"`
}

// SampleEntry holds a single Teams team entry
type SampleEntry struct {
	// ID is the Teams team identifier
	ID string `json:"id"`
	// DisplayName is the team display name
	DisplayName string `json:"displayName"`
}

// TeamsSample collects a sample of joined Microsoft Teams
type TeamsSample struct {
	// Teams is the collected team sample
	Teams []SampleEntry `json:"teams"`
}

// MessageSend sends a Microsoft Teams channel message
type MessageSend struct {
	// TeamID is the target team identifier
	TeamID string `json:"teamId"`
	// ChannelID is the target channel identifier
	ChannelID string `json:"channelId"`
	// MessageID is the identifier of the created message
	MessageID string `json:"messageId"`
}

type channelMessageBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type channelMessageRequest struct {
	Body    channelMessageBody `json:"body"`
	Subject string             `json:"subject,omitempty"`
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

// Run executes the Microsoft Teams health check
func (HealthCheck) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	var profile ProfileResponse
	if err := c.GetJSON(ctx, "me", &profile); err != nil {
		return nil, fmt.Errorf("microsoftteams: graph /me failed: %w", err)
	}

	return jsonx.ToRawMessage(HealthCheck{
		ID:   profile.ID,
		Mail: profile.Mail,
	})
}

// Handle adapts teams sample to the generic operation registration boundary
func (t TeamsSample) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return t.Run(ctx, c)
	}
}

// Run collects a sample of joined Microsoft Teams
func (TeamsSample) Run(ctx context.Context, c *providerkit.AuthenticatedClient) (json.RawMessage, error) {
	var resp struct {
		Value []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	if err := c.GetJSON(ctx, "me/joinedTeams?$top=5", &resp); err != nil {
		return nil, fmt.Errorf("microsoftteams: graph joinedTeams failed: %w", err)
	}

	samples := make([]SampleEntry, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, SampleEntry{
			ID:          team.ID,
			DisplayName: team.DisplayName,
		})
	}

	return jsonx.ToRawMessage(TeamsSample{Teams: samples})
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

// Run sends a Microsoft Teams channel message via Microsoft Graph
func (MessageSend) Run(ctx context.Context, c *providerkit.AuthenticatedClient, cfg MessageOperationInput) (json.RawMessage, error) {
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

	payload := channelMessageRequest{
		Body: channelMessageBody{
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

	return jsonx.ToRawMessage(MessageSend{
		TeamID:    cfg.TeamID,
		ChannelID: cfg.ChannelID,
		MessageID: resp.ID,
	})
}
