package vercel

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck identifies the default health check operation
type HealthCheck struct{}

// ProjectsSample identifies the project sample collection operation
type ProjectsSample struct{}

var (
	DefinitionID            = types.NewDefinitionRef("def_01K0VERCEL00000000000000001")
	HealthDefaultOperation  = types.NewOperationRef[HealthCheck]("health.default")
	ProjectsSampleOperation = types.NewOperationRef[ProjectsSample]("projects.sample")
)

const Slug = "vercel"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label     string `json:"label,omitempty"     jsonschema:"title=Installation Label"`
	TeamID    string `json:"teamId,omitempty"    jsonschema:"title=Team ID"`
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
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

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
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "Vercel REST API client",
					Build:       buildVercelClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Vercel /v2/user to verify token and account",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        ProjectsSampleOperation.Name(),
					Description: "Collect a sample of Vercel projects for drift detection",
					Topic:       ProjectsSampleOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runProjectsSampleOperation,
				},
			},
		}, nil
	})
}
