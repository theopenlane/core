package slack

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
}

// messageOperationInput holds per-invocation parameters for the message.send operation
type messageOperationInput struct {
	Channel     string            `json:"channel"                jsonschema:"required,title=Channel"`
	Text        string            `json:"text,omitempty"         jsonschema:"title=Message Text"`
	Blocks      []json.RawMessage `json:"blocks,omitempty"       jsonschema:"title=Block Kit Payload"`
	ThreadTS    string            `json:"thread_ts,omitempty"    jsonschema:"title=Thread Timestamp"`
	UnfurlLinks *bool             `json:"unfurl_links,omitempty" jsonschema:"title=Unfurl Links"`
}

// def holds operator config for the Slack integration
type def struct {
	cfg Config
}

// Builder returns the Slack definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0SLACK000000000000000001",
				Slug:        "slack",
				Version:     "v1",
				Family:      "slack",
				DisplayName: "Slack",
				Description: "Integrate with Slack to verify workspace posture and send operational or compliance notifications to channels.",
				Category:    "collab",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/slack/overview",
				Labels:      map[string]string{"vendor": "slack"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Auth: &types.AuthRegistration{
				Start:    d.startInstallAuth,
				Complete: d.completeInstallAuth,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "Slack Web API client",
					Build:       buildSlackClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Call auth.test to ensure the Slack token is valid and scoped correctly",
					Topic:       gala.TopicName("integration.slack.health.default"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "team.inspect",
					Kind:        types.OperationKindCollect,
					Description: "Collect workspace metadata via team.info for posture analysis",
					Topic:       gala.TopicName("integration.slack.team.inspect"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runTeamInspectOperation,
				},
				{
					Name:         "message.send",
					Kind:         types.OperationKindSync,
					Description:  "Send a Slack message via chat.postMessage",
					Topic:        gala.TopicName("integration.slack.message.send"),
					Client:       "api",
					ConfigSchema: providerkit.SchemaFrom[messageOperationInput](),
					Handle:       runMessageSendOperation,
				},
			},
		}, nil
	})
}
