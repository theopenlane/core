package okta

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck identifies the default health check operation
type HealthCheck struct{}

// PoliciesCollect identifies the policy collection operation
type PoliciesCollect struct{}

var (
	DefinitionID             = types.NewDefinitionRef("def_01K0OKTA0000000000000000001")
	HealthDefaultOperation   = types.NewOperationRef[HealthCheck]("health.default")
	PoliciesCollectOperation = types.NewOperationRef[PoliciesCollect]("policies.collect")
)

const Slug = "okta"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label  string `json:"label,omitempty"  jsonschema:"title=Installation Label"`
	OrgURL string `json:"orgUrl,omitempty" jsonschema:"title=Org URL"`
}

// credential holds the Okta tenant credentials for one installation
type credential struct {
	OrgURL   string `json:"orgUrl"   jsonschema:"required,title=Org URL"`
	APIToken string `json:"apiToken" jsonschema:"required,title=API Token"`
}

// Builder returns the Okta definition builder
func Builder() definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Version:     "v1",
				Family:      "okta",
				DisplayName: "Okta",
				Description: "Collect Okta tenant and sign-on policy metadata for identity posture and access governance analysis.",
				Category:    "sso",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/okta/overview",
				Labels:      map[string]string{"vendor": "okta", "product": "identity"},
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
					Description: "Okta API client",
					Build:       buildOktaClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Okta user API to verify API token",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:        PoliciesCollectOperation.Name(),
					Description: "Collect sign-on policy metadata for posture analysis",
					Topic:       PoliciesCollectOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Handle:      runPoliciesCollectOperation,
				},
			},
		}, nil
	})
}
