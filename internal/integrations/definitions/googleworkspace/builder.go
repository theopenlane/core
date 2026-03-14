package googleworkspace

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label              string `json:"label,omitempty"                  jsonschema:"title=Installation Label"`
	AdminEmail         string `json:"adminEmail,omitempty"             jsonschema:"title=Admin Email"`
	CustomerID         string `json:"customerId,omitempty"             jsonschema:"title=Customer ID"`
	OrganizationalUnit string `json:"organizationalUnitPath,omitempty" jsonschema:"title=Organizational Unit Path"`
	IncludeSuspended   bool   `json:"includeSuspendedUsers,omitempty"  jsonschema:"title=Include Suspended Users"`
	EnableGroupSync    bool   `json:"enableGroupSync,omitempty"        jsonschema:"title=Sync Groups"`
}

// def holds operator config for the Google Workspace integration
type def struct {
	cfg Config
}

// Builder returns the Google Workspace definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0GWKSP000000000000000001",
				Slug:        "google_workspace",
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
				Start:    d.startInstallAuth,
				Complete: d.completeInstallAuth,
			},
			Clients: []types.ClientRegistration{
				{
					Name:        "api",
					Description: "Google Workspace Admin SDK client",
					Build:       buildWorkspaceClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Call Google Admin SDK users.list to verify the workspace token",
					Topic:       gala.TopicName("integration.google_workspace.health.default"),
					Client:      "api",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "directory.sync",
					Kind:        types.OperationKindSync,
					Description: "Collect Google Workspace directory users and emit directory account envelopes",
					Topic:       gala.TopicName("integration.google_workspace.directory.sync"),
					Client:      "api",
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
