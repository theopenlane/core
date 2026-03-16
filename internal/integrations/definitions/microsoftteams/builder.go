package microsoftteams

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck identifies the default health check operation
type HealthCheck struct{}

// TeamsSample identifies the teams sample collection operation
type TeamsSample struct{}

// MessageSend identifies the Teams message send operation
type MessageSend struct{}

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0MSTEAMS00000000000000001")
	HealthDefaultOperation = types.NewOperationRef[HealthCheck]("health.default")
	TeamsSampleOperation   = types.NewOperationRef[TeamsSample]("teams.sample")
	MessageSendOperation   = types.NewOperationRef[MessageSend]("message.send")
)

const Slug = "microsoft_teams"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label    string `json:"label,omitempty"    jsonschema:"title=Installation Label"`
	TenantID string `json:"tenantId,omitempty" jsonschema:"title=Tenant ID"`
}

// messageOperationInput holds per-invocation parameters for the message.send operation
type messageOperationInput struct {
	TeamID     string `json:"team_id"               jsonschema:"required,title=Team ID"`
	ChannelID  string `json:"channel_id"            jsonschema:"required,title=Channel ID"`
	Body       string `json:"body"                  jsonschema:"required,title=Message Body"`
	BodyFormat string `json:"body_format,omitempty" jsonschema:"title=Body Format"`
	Subject    string `json:"subject,omitempty"     jsonschema:"title=Subject"`
}

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

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
				Schema: providerkit.SchemaFrom[userInput](),
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
					Ref:         clientRef.ID(),
					Description: "Microsoft Graph API client",
					Build:       buildGraphClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Graph /me to verify Teams access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        TeamsSampleOperation.Name(),
					Description: "Collect a sample of joined teams for the user context",
					Topic:       TeamsSampleOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runTeamsSampleOperation,
				},
				{
					Name:         MessageSendOperation.Name(),
					Description:  "Send a Teams channel message via Microsoft Graph",
					Topic:        MessageSendOperation.Topic(Slug),
					ClientRef:    clientRef.ID(),
					ConfigSchema: providerkit.SchemaFrom[messageOperationInput](),
					Handle:       runMessageSendOperation,
				},
			},
		}, nil
	})
}
