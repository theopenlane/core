package microsoftteams

import (
	"context"
	"encoding/json"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// MessageSendOp is the operation ref for the Microsoft Teams message send operation
var MessageSendOp = types.OperationPayloadOf[MessageSendOperation]().Handles(sendMessage)

// MessageSendOperation holds per-invocation parameters for the message.send operation
type MessageSendOperation struct {
	operations.MessageContent
	// BodyFormat controls the content type: text or html
	BodyFormat string `json:"bodyFormat,omitempty" jsonschema:"title=Body Format,enum=text,enum=html"`
}

// sendMessage sends a Microsoft Teams channel message via Microsoft Graph
func sendMessage(ctx context.Context, req types.OperationRequest, c *msgraphsdk.GraphServiceClient, cfg MessageSendOperation) (json.RawMessage, error) {
	if err := cfg.ResolveTemplate(ctx, req, &cfg); err != nil {
		return nil, err
	}

	if cfg.TeamID == "" || cfg.Channel == "" {
		return nil, ErrChannelMissing
	}

	if cfg.Text == "" {
		return nil, ErrMessageEmpty
	}

	contentType := models.TEXT_BODYTYPE
	if cfg.BodyFormat == "html" {
		contentType = models.HTML_BODYTYPE
	}

	msgBody := models.NewItemBody()
	msgBody.SetContent(&cfg.Text)
	msgBody.SetContentType(&contentType)

	msg := models.NewChatMessage()
	msg.SetBody(msgBody)

	if cfg.Subject != "" {
		msg.SetSubject(&cfg.Subject)
	}

	result, err := c.Teams().ByTeamId(cfg.TeamID).Channels().ByChannelId(cfg.Channel).Messages().Post(ctx, msg, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("team_id", cfg.TeamID).Str("channel", cfg.Channel).Msg("microsoftteams: channel message send failed")
		return nil, ErrChannelMessageSendFailed
	}

	return providerkit.EncodeResult(operations.NewMessageSendResult([]operations.MessageDelivery{{
		TeamID:  cfg.TeamID,
		Channel: cfg.Channel,
		ID:      lo.FromPtr(result.GetId()),
	}}), ErrResultEncode)
}
