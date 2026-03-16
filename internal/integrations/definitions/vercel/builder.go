package vercel

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// Label is the user-defined display label for the installation
	Label string `json:"label,omitempty" jsonschema:"title=Installation Label"`
	// TeamID scopes collection to a specific Vercel team
	TeamID string `json:"teamId,omitempty" jsonschema:"title=Team ID"`
	// ProjectID scopes collection to a specific Vercel project
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
}

// credential holds the Vercel API credentials for one installation
type credential struct {
	APIToken  string `json:"apiToken"            jsonschema:"required,title=API Token"`
	TeamID    string `json:"teamId,omitempty"    jsonschema:"title=Team ID"`
	ProjectID string `json:"projectId,omitempty" jsonschema:"title=Project ID"`
}

// Builder returns the Vercel definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
				Family:      "vercel",
				DisplayName: "Vercel",
				Description: "Collect Vercel project and deployment context to support devops posture and drift detection workflows.",
				Category:    "devops",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/vercel/overview",
				Labels:      map[string]string{"vendor": "vercel", "product": "deployment"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         VercelClient.ID(),
					Description: "Vercel REST API client",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Vercel /v2/user to verify token and account",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   VercelClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        ProjectsSampleOperation.Name(),
					Description: "Collect a sample of Vercel projects for drift detection",
					Topic:       ProjectsSampleOperation.Topic(Slug),
					ClientRef:   VercelClient.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      ProjectsSample{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
