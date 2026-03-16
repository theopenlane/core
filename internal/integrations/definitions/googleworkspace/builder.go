package googleworkspace

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID           = types.NewDefinitionRef("def_01K0GWKSP000000000000000001")
	HealthDefaultOperation = types.NewOperationRef[struct{}]("health.default")
	DirectorySyncOperation = types.NewOperationRef[struct{}]("directory.sync")
)

const Slug = "google_workspace"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label              string `json:"label,omitempty"                  jsonschema:"title=Installation Label"`
	AdminEmail         string `json:"adminEmail,omitempty"             jsonschema:"title=Admin Email"`
	CustomerID         string `json:"customerId,omitempty"             jsonschema:"title=Customer ID"`
	OrganizationalUnit string `json:"organizationalUnitPath,omitempty" jsonschema:"title=Organizational Unit Path"`
	IncludeSuspended   bool   `json:"includeSuspendedUsers,omitempty"  jsonschema:"title=Include Suspended Users"`
	EnableGroupSync    bool   `json:"enableGroupSync,omitempty"        jsonschema:"title=Sync Groups"`
}

// Builder returns the Google Workspace definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

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
				Schema: providerkit.SchemaFrom[userInput](),
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
					Ref:         clientRef.ID(),
					Description: "Google Workspace Admin SDK client",
					Build:       buildWorkspaceClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Google Admin SDK users.list to verify the workspace token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        DirectorySyncOperation.Name(),
					Description: "Collect Google Workspace directory users and emit directory account envelopes",
					Topic:       DirectorySyncOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runDirectorySyncOperation,
				},
			},
			Mappings: []types.MappingRegistration{
				{Schema: "directory_account", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
			},
		}, nil
	})
}
