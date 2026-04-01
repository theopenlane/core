package azureentraid

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Azure EntraID definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "azure",
				DisplayName: "Azure EntraID",
				Description: "Connect to Microsoft Graph to validate tenant access and inspect Azure Entra ID organization metadata.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_entra_id/overview",
				Tags:        []string{"directory-sync"},
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
					Ref:         entraTenantCredential.ID(),
					Name:        "Azure Entra ID Credential",
					Description: "OAuth credential used to access Microsoft Graph for Entra ID directory data.",
					Schema:      entraTenantSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       entraTenantCredential.ID(),
					Name:                "Azure Entra ID OAuth",
					Description:         "Connect your Azure Entra ID tenant using admin consent.",
					CredentialRefs:      []types.CredentialSlotID{entraTenantCredential.ID()},
					ClientRefs:          []types.ClientID{entraCredential.ID(), entraClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
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
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: entraTenantCredential.ID(),
						Description:   "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Azure Entra ID enterprise applications.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            entraCredential.ID(),
					CredentialRefs: []types.CredentialSlotID{entraTenantCredential.ID()},
					Description:    "Azure client credentials token credential for auth verification",
					Build:          CredentialClient{cfg: cfg}.Build,
				},
				{
					Ref:            entraClient.ID(),
					CredentialRefs: []types.CredentialSlotID{entraTenantCredential.ID()},
					Description:    "Microsoft Graph service client for directory operations",
					Build:          GraphClient{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Verify Azure client credentials can acquire a token against Microsoft Graph",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    entraCredential.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect Azure Entra ID users, groups, and memberships as directory accounts",
					Topic:        definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:    entraClient.ID(),
					ConfigSchema: directorySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
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
