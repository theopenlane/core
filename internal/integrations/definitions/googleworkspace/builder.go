package googleworkspace

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Google Workspace definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "google",
				DisplayName: "Google Workspace",
				Description: "Collect Google Workspace directory and identity metadata to support account hygiene and compliance posture checks.",
				Category:    "identity",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/google_workspace/overview",
				Labels:      map[string]string{"vendor": "google", "product": "workspace"},
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
					Ref:         workspaceCredential,
					Name:        "Google Workspace Credential",
					Description: "Auth-managed credential slot used by the Google Workspace client in this definition.",
				},
			},
			Auth: &types.AuthRegistration{
				StartPath:    types.DefaultAuthStartPath,
				CallbackPath: types.DefaultAuthCompletePath,
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
					TokenURL:    "https://oauth2.googleapis.com/token",
					RedirectURI: cfg.RedirectURL,
					Scopes: []string{
						"https://www.googleapis.com/auth/admin.directory.user.readonly",
						"https://www.googleapis.com/auth/admin.directory.group.readonly",
						"https://www.googleapis.com/auth/apps.groups.migration",
					},
					AuthParams: map[string]string{
						"access_type": "offline",
						"prompt":      "consent",
					},
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            WorkspaceClient.ID(),
					CredentialRefs: []types.CredentialRef{workspaceCredential},
					Description:    "Google Workspace Admin SDK client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Google Admin SDK users.list to verify the workspace token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   WorkspaceClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect Google Workspace directory users, groups, and memberships and emit directory ingest envelopes",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   WorkspaceClient.ID(),
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
