package slack

import (
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
}

// MessageOperationInput holds per-invocation parameters for the message.send operation
type MessageOperationInput struct {
	// Channel is the Slack channel identifier to post to
	Channel string `json:"channel" jsonschema:"required,title=Channel"`
	// Text is the plain-text message content
	Text string `json:"text,omitempty" jsonschema:"title=Message Text"`
	// Blocks is a Block Kit payload for rich messages
	Blocks []json.RawMessage `json:"blocks,omitempty" jsonschema:"title=Block Kit Payload"`
	// ThreadTS is the thread timestamp for replies
	ThreadTS string `json:"thread_ts,omitempty" jsonschema:"title=Thread Timestamp"`
	// UnfurlLinks controls whether links are unfurled
	UnfurlLinks *bool `json:"unfurl_links,omitempty" jsonschema:"title=Unfurl Links"`
}

// Builder returns the Slack definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
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
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Auth: &types.AuthRegistration{
				StartPath:    "/v1/integrations/oauth/start",
				CallbackPath: "/v1/integrations/oauth/callback",
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					AuthURL:     "https://slack.com/oauth/v2/authorize",
					TokenURL:    "https://slack.com/api/oauth.v2.access",
					RedirectURI: cfg.RedirectURL,
					Scopes: []string{
						"chat:write",
						"chat:write.public",
						"chat:write.customize",
						"team:read",
						"users:read",
					},
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         SlackClient.ID(),
					Description: "Slack Web API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call auth.test to ensure the Slack token is valid and scoped correctly",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SlackClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        TeamInspectOperation.Name(),
					Description: "Collect workspace metadata via team.info for posture analysis",
					Topic:       TeamInspectOperation.Topic(Slug),
					ClientRef:   SlackClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      TeamInspect{}.Handle(Client{}),
				},
				{
					Name:         MessageSendOperation.Name(),
					Description:  "Send a Slack message via chat.postMessage",
					Topic:        MessageSendOperation.Topic(Slug),
					ClientRef:    SlackClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[MessageOperationInput](),
					Handle:       MessageSend{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
