package azureentraid

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Azure EntraID definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "azure",
				DisplayName: "Azure EntraID",
				Description: "Connect to Microsoft Graph to validate tenant access and inspect Azure Entra ID organization metadata.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_entra_id/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "EntraID"},
				Active:      false,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: providerkit.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         entraTenantCredential,
					Name:        "Azure Entra ID Credential",
					Description: "Credential slot shared by the Azure Entra ID clients in this definition.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       entraTenantCredential,
					Name:                "Azure Entra ID OAuth",
					Description:         "Authenticate with Microsoft Graph using delegated tenant admin consent.",
					CredentialRefs:      []types.CredentialRef{entraTenantCredential},
					ClientRefs:          []types.ClientID{EntraCredential.ID(), EntraClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[entraIDCred]{
						CredentialRef: entraTenantCredential,
						Config: auth.OAuthConfig{
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
							TokenURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"openid",
								"profile",
								"offline_access",
								graphScope,
							},
							AuthParams: map[string]string{
								"prompt": "admin_consent",
							},
						},
						Material: func(material auth.OAuthMaterial) (entraIDCred, error) {
							if material.Claims == nil {
								return entraIDCred{}, ErrTenantIDNotFound
							}

							value, ok := material.Claims["tid"]
							if !ok {
								return entraIDCred{}, ErrTenantIDNotFound
							}

							tenantID, ok := value.(string)
							if !ok || tenantID == "" {
								return entraIDCred{}, ErrTenantIDNotFound
							}

							return entraIDCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
								TenantID:     tenantID,
							}, nil
						},
						TokenView: func(cred entraIDCred) (*types.TokenView, error) {
							return &types.TokenView{
								AccessToken: cred.AccessToken,
								ExpiresAt:   cred.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
						DecodeCredentialError: ErrCredentialDecode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: entraTenantCredential,
						Name:          "Disconnect Azure Entra ID OAuth",
						Description:   "Remove the persisted Azure Entra ID OAuth credential and disconnect this installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            EntraCredential.ID(),
					CredentialRefs: []types.CredentialRef{entraTenantCredential},
					Description:    "Azure client credentials token credential for auth verification",
					Build:          CredentialClient{cfg: cfg}.Build,
				},
				{
					Ref:            EntraClient.ID(),
					CredentialRefs: []types.CredentialRef{entraTenantCredential},
					Description:    "Microsoft Graph service client for directory operations",
					Build:          GraphClient{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Verify Azure client credentials can acquire a token against Microsoft Graph",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   EntraCredential.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        DirectoryInspectOperation.Name(),
					Description: "Collect basic tenant metadata via Microsoft Graph",
					Topic:       DirectoryInspectOperation.Topic(Slug),
					ClientRef:   EntraClient.ID(),
					Handle:      DirectoryInspect{}.Handle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect Azure Entra ID users, groups, and memberships as directory accounts",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   EntraClient.ID(),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
						},
					},
					IngestHandle: DirectorySync{}.IngestHandle(),
				},
			},
			Mappings: entraIDMappings(),
		}, nil
	})
}
