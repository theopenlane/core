package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// testMessageBodyFormat is the fixed text posted by a messaging integration's test message operation
const testMessageBodyFormat = "This is a test message from Openlane. Your %s integration is working."

// messageConfig is a pointer to a message send operation config embedding MessageContent
type messageConfig[Config any] interface {
	*Config
	SetMessageContent(content MessageContent)
}

// TestMessageSendOperation holds the target the console's shared messaging picker sends for a test message
type TestMessageSendOperation struct {
	// TeamID is the workspace or team the channel belongs to
	TeamID string `json:"teamId,omitempty" jsonschema:"title=Team ID"`
	// ChannelID is the target channel identifier
	ChannelID string `json:"channelId" jsonschema:"required,title=Channel ID"`
}

// NewTestMessageOperation builds the inline operation that posts a fixed test message to the picked channel through the integration's message send handler
func NewTestMessageOperation[C any, Config any, P messageConfig[Config]](send func(context.Context, types.OperationRequest, C, Config) (json.RawMessage, error)) types.OperationRef[TestMessageSendOperation] {
	return types.OperationPayloadOf[TestMessageSendOperation]().
		Handles(testMessage[C, Config, P]{send: send}.run).
		Policy(types.ExecutionPolicy{Inline: true})
}

// testMessage binds an integration's message send handler to the test message operation
type testMessage[C any, Config any, P messageConfig[Config]] struct {
	// send is the integration's message send handler
	send func(context.Context, types.OperationRequest, C, Config) (json.RawMessage, error)
}

// run posts the fixed test body, naming the installation, to the picked channel
func (t testMessage[C, Config, P]) run(ctx context.Context, req types.OperationRequest, client C, target TestMessageSendOperation) (json.RawMessage, error) {
	var cfg Config

	P(&cfg).SetMessageContent(MessageContent{
		TeamID:  target.TeamID,
		Channel: target.ChannelID,
		Text:    fmt.Sprintf(testMessageBodyFormat, req.Integration.Name),
	})

	return t.send(ctx, req, client, cfg)
}
