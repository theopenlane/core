package microsoftteams

import (
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrationsv2/definition"
	"github.com/theopenlane/core/internal/integrationsv2/types"
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

var (
	definitionSpec              = types.DefinitionSpec{
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
	}

	configSchema                = providerkit.SchemaFrom[Config]()
	userInputSchema             = providerkit.SchemaFrom[userInput]()
	messageOperationInputSchema = providerkit.SchemaFrom[messageOperationInput]()
)

// def implements definition.Assembler for the Microsoft Teams integration
type def struct {
	cfg Config
}

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.FromAssembler(&def{cfg: cfg})
}

func (d *def) Spec() types.DefinitionSpec { return definitionSpec }

func (d *def) OperatorConfig() *types.OperatorConfigRegistration {
	return &types.OperatorConfigRegistration{Schema: configSchema}
}

func (d *def) UserInput() *types.UserInputRegistration {
	return &types.UserInputRegistration{Schema: userInputSchema}
}

func (d *def) Credentials() *types.CredentialRegistration { return nil }

func (d *def) Auth() *types.AuthRegistration {
	return &types.AuthRegistration{
		Start:    startInstallAuth,
		Complete: completeInstallAuth,
	}
}

func (d *def) Clients() []types.ClientRegistration {
	return []types.ClientRegistration{
		{
			Name:        "api",
			Description: "Microsoft Graph API client",
			Build:       buildGraphClient,
		},
	}
}

func (d *def) Operations() []types.OperationRegistration {
	return []types.OperationRegistration{
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
			ConfigSchema: messageOperationInputSchema,
			Handle:       runMessageSendOperation,
		},
	}
}

func (d *def) Mappings() []types.MappingRegistration { return nil }
func (d *def) Webhooks() []types.WebhookRegistration { return nil }
