package googleworkspace

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Google Workspace definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "google",
				DisplayName: "Google Workspace",
				Description: "Collect Google Workspace directory and identity metadata to support account hygiene and compliance posture checks.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/google_workspace/overview",
				Tags:        []string{"directory-sync"},
				Active:      true,
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
					Ref:         workspaceCredential.ID(),
					Name:        "Google Workspace Credential",
					Description: "OAuth credential used to access Google Workspace directory data.",
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       workspaceCredential.ID(),
					Name:                "Google Workspace OAuth",
					Description:         "Connect your Google Workspace domain using OAuth.",
					CredentialRefs:      []types.CredentialSlotID{workspaceCredential.ID()},
					ClientRefs:          []types.ClientID{workspaceClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Auth: auth.OAuthRegistration(auth.OAuthRegistrationOptions[googleWorkspaceCred]{
						CredentialRef: workspaceCredential,
						Config: auth.OAuthConfig{
							ClientID:     cfg.ClientID,
							ClientSecret: cfg.ClientSecret,
							AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
							TokenURL:     "https://oauth2.googleapis.com/token",
							RedirectURL:  cfg.RedirectURL,
							Scopes: []string{
								"https://www.googleapis.com/auth/admin.directory.user.readonly",
								"https://www.googleapis.com/auth/admin.directory.group.readonly",
								"https://www.googleapis.com/auth/admin.directory.orgunit.readonly",
								"https://www.googleapis.com/auth/admin.directory.domain.readonly",
								"https://www.googleapis.com/auth/admin.directory.customer.readonly",
							},
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
						CredentialRef: workspaceCredential.ID(),
						Description:   "Removes the stored OAuth credential from Openlane. To fully revoke access, remove the Openlane app from your Google Workspace admin console under Security > API controls.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            workspaceClient.ID(),
					CredentialRefs: []types.CredentialSlotID{workspaceCredential.ID()},
					Description:    "Google Workspace Admin SDK client",
					Build:          Client{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Google Admin SDK users.list to verify the workspace token",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    workspaceClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: healthCheckSchema,
					Handle:       HealthCheck{}.Handle(),
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Collect Google Workspace directory users, groups, and memberships and emit directory ingest envelopes",
					Topic:        definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:    workspaceClient.ID(),
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
			Mappings: googleWorkspaceMappings(),
		}, nil
	})
}
