package googledrive

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Builder returns the Google Drive definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Google Drive",
				DisplayName: "Google Drive",
				Description: "Live, read-only integration with Google Drive for on-the-fly HTML export of Google Docs.",
				Category:    "document",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/google_drive/overview",
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
					Ref:         driveCredential.ID(),
					Name:        "Google Drive Credential",
					Description: "OAuth credential used to access Google Drive documents.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       driveCredential.ID(),
					Name:                "Google Drive OAuth",
					Description:         "Connect your Google account using OAuth to access Drive documents.",
					CredentialRefs:      []types.CredentialSlotID{driveCredential.ID()},
					ClientRefs:          []types.ClientID{driveClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[googleDriveCred]{
						CredentialRef: driveCredential,
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
							TokenURL:     "https://oauth2.googleapis.com/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
							AuthParams: map[string]string{
								"access_type": "offline",
								"prompt":      "consent",
							},
						},
						Material: func(material auth.OAuthMaterial) (googleDriveCred, error) {
							return googleDriveCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: driveCredential.ID(),
						Description:   "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Google account under Security > Third-party access.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            driveClient.ID(),
					CredentialRefs: []types.CredentialSlotID{driveCredential.ID()},
					Description:    "Google Drive API client",
					Build:          Client{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Google Drive about.get to verify the drive token",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    driveClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         documentExportOperation.Name(),
					Description:  "Export a Google Doc as HTML via the Drive files.export endpoint",
					Topic:        definitionID.OperationTopic(documentExportOperation.Name()),
					ClientRef:    driveClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: documentExportSchema,
					Handle:       Handle(),
				},
				{
					Name:        folderSyncOperation.Name(),
					Description: "List Google Docs in the configured folder and emit policy ingest envelopes",
					Topic:       definitionID.OperationTopic(folderSyncOperation.Name()),
					ClientRef:   driveClient.ID(),
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaInternalPolicy,
						},
					},
					IngestHandle: FolderSync{}.IngestHandle(),
					ConfigSchema: folderSyncSchema,
					Schedule:     gala.NewFullFetchSchedule(),
				},
			},
			Mappings: googleDriveMappings(),
		}, nil
	})
}
