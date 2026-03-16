package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
	// TenantID is the Microsoft tenant identifier
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// MessageOperationInput holds per-invocation parameters for the message.send operation
type MessageOperationInput struct {
	// TeamID is the target Microsoft Teams team identifier
	TeamID string `json:"team_id" jsonschema:"required,title=Team ID"`
	// ChannelID is the target Teams channel identifier
	ChannelID string `json:"channel_id" jsonschema:"required,title=Channel ID"`
	// Body is the message body content
	Body string `json:"body" jsonschema:"required,title=Message Body"`
	// BodyFormat controls the content type (text or html)
	BodyFormat string `json:"body_format,omitempty" jsonschema:"title=Body Format"`
	// Subject is an optional message subject
	Subject string `json:"subject,omitempty" jsonschema:"title=Subject"`
}

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
				Family:      "microsoft",
				DisplayName: "Microsoft Teams",
				Description: "Integrate with Microsoft Teams to collect collaboration metadata and send notification messages through Microsoft Graph.",
				Category:    "collab",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/microsoft_teams/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "teams"},
				Active:      false,
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
					AuthURL:     teamsAuthURL,
					TokenURL:    teamsTokenURL,
					RedirectURI: cfg.RedirectURL,
					Scopes:      teamsScopes,
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         TeamsClient.ID(),
					Description: "Microsoft Graph API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Graph /me to verify Teams access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   TeamsClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        TeamsSampleOperation.Name(),
					Description: "Collect a sample of joined teams for the user context",
					Topic:       TeamsSampleOperation.Topic(Slug),
					ClientRef:   TeamsClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      TeamsSample{}.Handle(Client{}),
				},
				{
					Name:         MessageSendOperation.Name(),
					Description:  "Send a Teams channel message via Microsoft Graph",
					Topic:        MessageSendOperation.Topic(Slug),
					ClientRef:    TeamsClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[MessageOperationInput](),
					Handle:       MessageSend{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
