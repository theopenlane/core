package slack

import (
	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0SLACK000000000000000001")
	SlackClient            = types.NewClientRef[*slackgo.Client]()
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	TeamInspectOperation   = types.NewOperationRef[TeamInspect]("team.inspect")
	MessageSendOperation   = types.NewOperationRef[MessageSend]("message.send")
)

const Slug = "slack"
