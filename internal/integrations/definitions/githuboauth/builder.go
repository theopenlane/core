package githuboauth

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label            string `json:"label,omitempty"            jsonschema:"title=Installation Label"`
	RepositoryFilter string `json:"repositoryFilter,omitempty" jsonschema:"title=Repository Filter Expression"`
	SecurityOnly     bool   `json:"securityOnly,omitempty"     jsonschema:"title=Collect Security Signals Only"`
}

// def holds operator config for the GitHub OAuth integration
type def struct {
	cfg Config
}

// Builder returns the GitHub OAuth definition builder with the supplied operator config applied
func Builder(cfg Config) definition.Builder {
	d := &def{cfg: cfg}
	return definition.BuilderFunc(func(_ context.Context) (types.Definition, error) {
		return types.Definition{
			Spec: types.DefinitionSpec{
				ID:          "def_01K0GHOAUTH0000000000000001",
				Slug:        "github_oauth",
				Version:     "v1",
				Family:      "github",
				DisplayName: "GitHub",
				Description: "Collect GitHub repository metadata and security alerts (Dependabot, code scanning, and secret scanning) for exposure management.",
				Category:    "code",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/github/overview",
				Labels:      map[string]string{"vendor": "github"},
				Active:      true,
				Visible:     false,
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
					Name:        "rest",
					Description: "GitHub REST API client",
					Build:       buildRESTClient,
				},
				{
					Name:        "graphql",
					Description: "GitHub GraphQL API client",
					Build:       buildGraphQLClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Kind:        types.OperationKindHealth,
					Description: "Validate GitHub OAuth token by calling the /user endpoint",
					Topic:       gala.TopicName("integration.github_oauth.health.default"),
					Client:      "rest",
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        "repository.sync",
					Kind:        types.OperationKindSync,
					Description: "Collect repository metadata for the authenticated account",
					Topic:       gala.TopicName("integration.github_oauth.repository.sync"),
					Client:      "rest",
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runRepositorySyncOperation,
				},
				{
					Name:        "vulnerability.collect",
					Kind:        types.OperationKindCollect,
					Description: "Collect GitHub vulnerability alerts (Dependabot, code scanning, secret scanning) for accessible repositories",
					Topic:       gala.TopicName("integration.github_oauth.vulnerability.collect"),
					Client:      "rest",
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runVulnerabilityCollectOperation,
				},
			},
			Mappings: []types.MappingRegistration{
				{Schema: "repository", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
				{Schema: "vulnerability", Variant: "default", Spec: types.MappingOverride{Version: "v1"}},
			},
		}, nil
	})
}
