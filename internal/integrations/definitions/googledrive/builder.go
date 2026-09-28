package googledrive

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the Google Drive definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		installation := installationRef(cfg)

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Google Drive",
				DisplayName: "Google Drive",
				Description: "Live, read-only integration with Google Drive for on-the-fly HTML export of Google Docs.",
				Category:    "document",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/google_drive",
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
				driveCredential.Registration(types.CredentialRegistration{
					Name:        "Google Drive Credential",
					Description: "OAuth credential used to access Google Drive documents.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				driveConnection.Registration(types.ConnectionRegistration{
					Name:           "Google Drive OAuth",
					Description:    "Connect your Google account using OAuth to access Drive documents.",
					CredentialRefs: []types.CredentialSlotID{driveCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: driveClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
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
				}),
			},
			Clients: []types.ClientRegistration{
				driveClient.Registration(Client{cfg: cfg}.Build, types.ClientRegistration{
					Description: "Google Drive API client",
				}),
			},
			Operations: []types.OperationRegistration{
				documentExportOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Export a Google Doc as HTML via the Drive files.export endpoint",
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle:      Handle(),
				}),
				folderSyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "List Google Docs in the configured folder and emit policy ingest envelopes",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaInternalPolicy.Name,
						},
					},
					IngestHandle: FolderSync{}.IngestHandle(),
					Schedule:     gala.NewFullFetchSchedule(),
				}),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaInternalPolicy.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprInternalPolicy,
					},
				},
			},
		}, nil
	})
}
