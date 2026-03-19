package slack

import (
	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	// DefinitionID is the stable identifier for the Slack integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0SLACK000000000000000001")
	// SlackClient is the client ref for the Slack Web API client used by this definition
	SlackClient = types.NewClientRef[*slackgo.Client]()
	// HealthDefaultOperation is the operation ref for the Slack health check
	HealthDefaultOperation = types.NewOperationRef[HealthCheck](types.HealthDefaultOperation)
	// TeamInspectOperation is the operation ref for the Slack team inspect operation
	TeamInspectOperation = types.NewOperationRef[TeamInspect]("team.inspect")
	// ChannelsListOperation is the operation ref for the Slack channels list operation
	ChannelsListOperation = types.NewOperationRef[ChannelsListOperationInput]("channels.list")
	// MessageSendOperation is the operation ref for the Slack message send operation
	MessageSendOperation = types.NewOperationRef[MessageOperationInput]("message.send")
)

// Slug is the unique identifier for the Slack integration
const Slug = "slack"

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression applied to imported records before ingest."`
}
