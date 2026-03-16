package slack

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0SLACK000000000000000001")
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	TeamInspectOperation   = types.NewOperationRef[struct{}]("team.inspect")
	MessageSendOperation   = types.NewOperationRef[struct{}]("message.send")
)

const Slug = "slack"

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

// Builder returns the Slack definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
				StartPath:    "/v1/integrations/oauth/start",
				CallbackPath: "/v1/integrations/oauth/callback",
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					AuthURL:     slackAuthURL,
					TokenURL:    slackTokenURL,
					RedirectURI: cfg.RedirectURL,
					Scopes:      slackScopes,
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "Slack Web API client",
					Build:       buildSlackClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call auth.test to ensure the Slack token is valid and scoped correctly",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        TeamInspectOperation.Name(),
					Description: "Collect workspace metadata via team.info for posture analysis",
					Topic:       TeamInspectOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runTeamInspectOperation,
				},
				{
					Name:         MessageSendOperation.Name(),
					Description:  "Send a Slack message via chat.postMessage",
					Topic:        MessageSendOperation.Topic(Slug),
					ClientRef:    clientRef.ID(),
					ConfigSchema: providerkit.SchemaFrom[messageOperationInput](),
					Handle:       runMessageSendOperation,
				},
			},
		}, nil
	})
}
