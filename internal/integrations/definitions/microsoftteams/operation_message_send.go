package microsoftteams

import (
	"context"
	"encoding/json"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// MessageOperationInput holds per-invocation parameters for the message.send operation
type MessageOperationInput struct {
	// TeamID is the target Microsoft Teams team identifier
	TeamID string `json:"team_id" jsonschema:"required,title=Team ID"`
	// ChannelID is the target Teams channel identifier
	ChannelID string `json:"channel_id" jsonschema:"required,title=Channel ID"`
	// Body is the message body content
	Body string `json:"body" jsonschema:"required,title=Message Body"`
	// BodyFormat controls the content type: text or html
	BodyFormat string `json:"body_format,omitempty" jsonschema:"title=Body Format,enum=text,enum=html"`
	// Subject is an optional message subject
	Subject string `json:"subject,omitempty" jsonschema:"title=Subject"`
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

// Handle adapts message send to the generic operation registration boundary
func (m MessageSend) Handle() types.OperationHandler {
	return providerkit.OperationWithClientConfig(TeamsClient, MessageSendOperation, ErrOperationConfigInvalid, m.Run)
}

// Run sends a Microsoft Teams channel message via Microsoft Graph
func (MessageSend) Run(ctx context.Context, c *msgraphsdk.GraphServiceClient, cfg MessageOperationInput) (json.RawMessage, error) {
	if cfg.TeamID == "" || cfg.ChannelID == "" {
		return nil, ErrChannelMissing
	}

	if cfg.Body == "" {
		return nil, ErrMessageEmpty
	}

	contentType := models.TEXT_BODYTYPE
	if cfg.BodyFormat == "html" {
		contentType = models.HTML_BODYTYPE
	}

	msgBody := models.NewItemBody()
	msgBody.SetContent(&cfg.Body)
	msgBody.SetContentType(&contentType)

	msg := models.NewChatMessage()
	msg.SetBody(msgBody)

	if cfg.Subject != "" {
		msg.SetSubject(&cfg.Subject)
	}

	result, err := c.Teams().ByTeamId(cfg.TeamID).Channels().ByChannelId(cfg.ChannelID).Messages().Post(ctx, msg, nil)
	if err != nil {
		return nil, ErrChannelMessageSendFailed
	}

	return providerkit.EncodeResult(MessageSend{
		TeamID:    cfg.TeamID,
		ChannelID: cfg.ChannelID,
		MessageID: lo.FromPtr(result.GetId()),
	}, ErrResultEncode)
}
