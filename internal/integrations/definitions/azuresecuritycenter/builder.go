package azuresecuritycenter

import (
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Azure Security Center definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
				Family:      "azure",
				DisplayName: "Microsoft Defender for Cloud",
				Description: "Collect Microsoft Defender for Cloud pricing and plan metadata from an Azure subscription for security posture visibility.",
				Category:    "compliance",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "defender-for-cloud"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         SecurityCenterClient.ID(),
					Description: "Azure management API client for Defender for Cloud",
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Azure Security Center pricings API to verify access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SecurityCenterClient.ID(),
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:        SecurityPricingOverviewOperation.Name(),
					Description: "Collect plan and pricing metadata for Microsoft Defender for Cloud",
					Topic:       SecurityPricingOverviewOperation.Topic(Slug),
					ClientRef:   SecurityCenterClient.ID(),
					Handle:      SecurityPricingOverview{}.Handle(Client{}),
				},
			},
		}, nil
	})
}
