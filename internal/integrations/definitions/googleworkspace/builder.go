package googleworkspace

import (
	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

var directorySyncScopes = []string{
	"https://www.googleapis.com/auth/admin.directory.user.readonly",
	"https://www.googleapis.com/auth/admin.directory.group.readonly",
	"https://www.googleapis.com/auth/admin.directory.orgunit.readonly",
	"https://www.googleapis.com/auth/admin.directory.domain.readonly",
	"https://www.googleapis.com/auth/admin.directory.customer.readonly",
}

// Builder returns the Google Workspace definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		installation := installationRef(cfg)

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Google Workspace",
				DisplayName: "Google Workspace",
				Description: "Collect Google Workspace directory and identity metadata to support account hygiene and compliance posture checks.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/google_workspace",
				Tags:        []string{"directory"},
				Active:      true,
				Visible:     true,
			},
			OperatorConfig: &types.OperatorConfigRegistration{
				Schema: jsonx.SchemaFrom[Config](),
			},
			UserInput:    userInput.Registration(),
			HealthCheck:  workspaceClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				workspaceCredential.Registration(types.CredentialRegistration{
					Name:        "Google Workspace Credential",
					Description: "OAuth credential used to access Google Workspace directory data.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				workspaceConnection.Registration(types.ConnectionRegistration{
					Name:        "Google Workspace OAuth",
					Description: "Connect your Google Workspace domain using OAuth.",
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[googleWorkspaceCred]{
						CredentialRef: workspaceCredential,
						Config: auth.OAuthConfig{ //nolint:gosec
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
							TokenURL:     "https://oauth2.googleapis.com/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes:       directorySyncScopes,
							AuthParams: map[string]string{
								"access_type": "offline",
								"prompt":      "consent",
							},
						},
						Material: func(material auth.OAuthMaterial) (googleWorkspaceCred, error) {
							return googleWorkspaceCred{
								AccessToken:  material.AccessToken,
								RefreshToken: material.RefreshToken,
								Expiry:       material.Expiry,
							}, nil
						},
						EncodeCredentialError: ErrCredentialEncode,
					}),
					Disconnect: &types.DisconnectRegistration{
						Description: "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Google Workspace admin console under Security > API controls.",
					},
				}),
			},
			Clients: []types.ClientRegistration{
				workspaceClient.Registration(Client{cfg: cfg}.Build, types.ClientRegistration{
					Description: "Google Workspace Admin SDK client",
				}),
			},
			Operations: []types.OperationRegistration{
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:         "Collect Google Workspace directory users, groups, and memberships and emit directory ingest envelopes",
					Policy:              types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Ingest:              providerkit.DirectoryIngestContracts(),
					SkipDefaultLookback: true,
					RequiredPermissions: directorySyncScopes,
					Schedule:            gala.NewFullFetchSchedule(),
				}),
			},
			Mappings: providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership),
		}, nil
	})
}
