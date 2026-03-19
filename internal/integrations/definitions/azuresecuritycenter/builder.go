package azuresecuritycenter

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
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
				Description: "Collect security assessment findings and vulnerability data from Microsoft Defender for Cloud across an Azure subscription.",
				Category:    "compliance",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center/overview",
				Labels:      map[string]string{"vendor": "microsoft", "product": "defender-for-cloud"},
				Active:      false,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         securityCenterCredential,
					Name:        "Azure Security Center Credential",
					Description: "Credential slot used by the Azure Security Center client in this definition.",
					Schema:      providerkit.SchemaFrom[CredentialSchema](),
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            SecurityCenterClient.ID(),
					CredentialRefs: []types.CredentialRef{securityCenterCredential},
					Description:    "Azure Security Center assessments and sub-assessments client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Call Azure Security Center assessments API to verify access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SecurityCenterClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:        AssessmentsCollectOperation.Name(),
					Description: "Collect unhealthy security posture assessment findings for vulnerability ingestion",
					Topic:       AssessmentsCollectOperation.Topic(Slug),
					ClientRef:   SecurityCenterClient.ID(),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: AssessmentsCollect{}.IngestHandle(),
				},
				{
					Name:        SubAssessmentsCollectOperation.Name(),
					Description: "Collect granular sub-assessment vulnerability findings (CVEs from container images, servers, and SQL checks)",
					Topic:       SubAssessmentsCollectOperation.Topic(Slug),
					ClientRef:   SecurityCenterClient.ID(),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: SubAssessmentsCollect{}.IngestHandle(),
				},
			},
			Mappings: azureSecurityCenterMappings(),
		}, nil
	})
}
