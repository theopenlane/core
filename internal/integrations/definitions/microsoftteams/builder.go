package microsoftteams

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

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

// def holds operator config for the Microsoft Teams integration
type def struct {
	cfg Config
}

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0MSTEAMS00000000000000001",
				Slug:        "microsoft_teams",
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
				Start:    d.startInstallAuth,
				Complete: d.completeInstallAuth,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "Microsoft Graph API client",
					Build:       buildGraphClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Call Graph /me to verify Teams access",
					Topic:       gala.TopicName("integration.microsoft_teams.health.default"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "teams.sample",
					Kind:        types.OperationKindCollect,
					Description: "Collect a sample of joined teams for the user context",
					Topic:       gala.TopicName("integration.microsoft_teams.teams.sample"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runTeamsSampleOperation,
				},
				{
					Name:         "message.send",
					Kind:         types.OperationKindSync,
					Description:  "Send a Teams channel message via Microsoft Graph",
					Topic:        gala.TopicName("integration.microsoft_teams.message.send"),
					Client:       "api",
					ConfigSchema: providerkit.SchemaFrom[messageOperationInput](),
					Handle:       runMessageSendOperation,
				},
			},
		}, nil
	})
}
