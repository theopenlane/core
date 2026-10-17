package microsoftteams

import (
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Microsoft Teams definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:          DefinitionID.ID(),
			Family:      "Microsoft",
			DisplayName: "Microsoft Teams",
			Description: "Send notification messages to Microsoft Teams channels via Microsoft Graph.",
			Category:    "collaboration",
			DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/microsoft_teams/",
			Tags:        []string{"messaging"},
			Active:      true,
			Visible:     true,
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			Installation: installation.Registration(),
			Connections: []types.Connector{
				oauthConnection.
					Name("Microsoft Teams OAuth").
					Description("Connect your Microsoft Teams workspace using OAuth.").
					Authenticates(auth.OAuthRegistration(auth.OAuthRegistrationOptions[teamsCred]{
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ApplicationID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      microsoft.AuthorizeURL(microsoft.OrganizationsTenant),
							TokenURL:     microsoft.TokenURL(microsoft.OrganizationsTenant),
							RedirectURL:  cfg.RedirectURL,
							Scopes:       teamsScopes,
						},
						Material: func(material auth.OAuthMaterial) (teamsCred, error) {
							return teamsCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					})).
					Provides(clientBuilder(cfg)).
					Verified(verify).
					Disconnects("Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Azure Entra ID enterprise applications.", nil),
			},
			Operations: []types.OperationRegistration{
				MessageSendOp.Description("Send a Teams channel message via Microsoft Graph").Registration(),
				operations.NewChannelsListOperation(listTeams, ErrResultEncode).Description("List the teams and channels the connected user can post to").Registration(),
				operations.NewTestMessageOperation(MessageSend{}.Run).Description("Send a test message to a Teams channel to confirm the integration works").Registration(),
			},
		}, nil
	})
}
