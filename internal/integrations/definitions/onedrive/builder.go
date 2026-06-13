package onedrive

import (
	"fmt"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const microsoftAuthBaseURL = "https://login.microsoftonline.com/%s/oauth2/v2.0"

// oauthTenant returns the tenant segment to use in Microsoft OAuth URLs.
// When DefaultTenant is configured it is used directly; otherwise "common" is returned.
func oauthTenant(cfg Config) string {
	if cfg.DefaultTenant != "" {
		return cfg.DefaultTenant
	}

	return "common"
}

// Builder returns the OneDrive definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Microsoft",
				DisplayName: "Microsoft OneDrive",
				Description: "Live, read-only integration with Microsoft OneDrive for document management and policy sync.",
				Category:    "document",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/onedrive/overview",
				Tags:        []string{"document"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         oneDriveCredential.ID(),
					Name:        "OneDrive Credential",
					Description: "OAuth credential used to access Microsoft OneDrive documents.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       oneDriveCredential.ID(),
					Name:                "OneDrive OAuth",
					Description:         "Connect your Microsoft account using OAuth to access OneDrive documents.",
					CredentialRefs:      []types.CredentialSlotID{oneDriveCredential.ID()},
					ClientRefs:          []types.ClientID{oneDriveClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[oneDriveCred]{
						CredentialRef: oneDriveCredential,
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      fmt.Sprintf(microsoftAuthBaseURL, oauthTenant(cfg)) + "/authorize",
							TokenURL:     fmt.Sprintf(microsoftAuthBaseURL, oauthTenant(cfg)) + "/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"https://graph.microsoft.com/Files.Read",
								"https://graph.microsoft.com/User.Read",
								"offline_access",
							},
						},
						Material: func(material auth.OAuthMaterial) (oneDriveCred, error) {
							return oneDriveCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: oneDriveCredential.ID(),
						Description:   "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Microsoft account under Account settings > Privacy.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            oneDriveClient.ID(),
					CredentialRefs: []types.CredentialSlotID{oneDriveCredential.ID()},
					Description:    "Microsoft OneDrive Graph API client",
					Build:          Client{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Microsoft Graph /me/drive to verify OneDrive access",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    oneDriveClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         documentExportOperation.Name(),
					Description:  "Download a OneDrive file and return its content",
					Topic:        definitionID.OperationTopic(documentExportOperation.Name()),
					ClientRef:    oneDriveClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: documentExportSchema,
					Handle:       Handle(),
				},
				{
					Name:        folderSyncOperation.Name(),
					Description: "List document files in the configured OneDrive folder and emit policy ingest envelopes",
					Topic:       definitionID.OperationTopic(folderSyncOperation.Name()),
					ClientRef:   oneDriveClient.ID(),
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaInternalPolicy,
						},
					},
					IngestHandle:      FolderSync{}.IngestHandle(),
					ConfigSchema:      folderSyncSchema,
					ReconcileSchedule: gala.NewFullFetchSchedule(),
				},
			},
			Mappings: oneDriveMappings(),
		}, nil
	})
}
