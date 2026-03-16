package googleworkspace

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
	// AdminEmail is the delegated admin email for impersonation
	AdminEmail string `json:"adminEmail,omitempty" jsonschema:"title=Admin Email"`
	// CustomerID is the Google Workspace customer identifier
	CustomerID string `json:"customerId,omitempty" jsonschema:"title=Customer ID"`
	// OrganizationalUnit limits collection to a specific org unit path
	OrganizationalUnit string `json:"organizationalUnitPath,omitempty" jsonschema:"title=Organizational Unit Path"`
	// IncludeSuspended controls whether suspended users are included
	IncludeSuspended bool `json:"includeSuspendedUsers,omitempty" jsonschema:"title=Include Suspended Users"`
	// EnableGroupSync controls whether group membership is collected
	EnableGroupSync bool `json:"enableGroupSync,omitempty" jsonschema:"title=Sync Groups"`
}

// Builder returns the Google Workspace definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
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
			Auth: &types.AuthRegistration{
				StartPath:    "/v1/integrations/oauth/start",
				CallbackPath: "/v1/integrations/oauth/callback",
				OAuth: &types.OAuthPublicConfig{
					ClientID:    cfg.ClientID,
					AuthURL:     googleAuthURL,
					TokenURL:    googleTokenURL,
					RedirectURI: cfg.RedirectURL,
					Scopes:      googleWorkspaceScopes,
					AuthParams:  googleWorkspaceAuthParams,
				},
				ClientSecret: cfg.ClientSecret,
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         WorkspaceClient.ID(),
					Description: "Google Workspace Admin SDK client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Google Admin SDK users.list to verify the workspace token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   WorkspaceClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:         DirectorySyncOperation.Name(),
					Description:  "Collect Google Workspace directory users, groups, and memberships and emit directory ingest envelopes",
					Topic:        DirectorySyncOperation.Topic(Slug),
					ClientRef:    WorkspaceClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[DirectorySyncConfig](),
					Policy:       types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
							EnsurePayloads: true,
						},
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
							EnsurePayloads: true,
						},
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
							EnsurePayloads: true,
						},
					},
					Handle: DirectorySync{}.Handle(Client{}),
				},
			},
			Mappings: googleWorkspaceMappings(),
		}, nil
	})
}
