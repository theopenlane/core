package azuresecuritycenter

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the Azure Security Center definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
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
					Ref:         securityCenterCredential.ID(),
					Name:        "Azure Security Center Credential",
					Description: "Credential slot used by the Azure Security Center client in this definition.",
					Schema:      securityCenterSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       securityCenterCredential.ID(),
					Name:                "Azure Service Principal",
					Description:         "Configure Microsoft Defender for Cloud access using an Azure service principal credential.",
					CredentialRefs:      []types.CredentialSlotID{securityCenterCredential.ID()},
					ClientRefs:          []types.ClientID{securityCenterClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: securityCenterCredential.ID(),
						Name:          "Disconnect Azure Service Principal",
						Description:   "Remove the persisted Azure service principal credential and disconnect this Microsoft Defender for Cloud installation from Openlane.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            securityCenterClient.ID(),
					CredentialRefs: []types.CredentialSlotID{securityCenterCredential.ID()},
					Description:    "Azure Security Center assessments and sub-assessments client",
					Build:          Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Call Azure Security Center assessments API to verify access",
					Topic:        types.OperationTopic(definitionID.ID(), healthCheckOperation.Name()),
					ClientRef:    securityCenterClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         assessmentsCollectOperation.Name(),
					Description:  "Collect unhealthy security posture assessment findings for vulnerability ingestion",
					Topic:        types.OperationTopic(definitionID.ID(), assessmentsCollectOperation.Name()),
					ClientRef:    securityCenterClient.ID(),
					ConfigSchema: assessmentsCollectSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: AssessmentsCollect{}.IngestHandle(),
				},
				{
					Name:         subAssessmentsCollectOperation.Name(),
					Description:  "Collect granular sub-assessment vulnerability findings (CVEs from container images, servers, and SQL checks)",
					Topic:        types.OperationTopic(definitionID.ID(), subAssessmentsCollectOperation.Name()),
					ClientRef:    securityCenterClient.ID(),
					ConfigSchema: subAssessmentsCollectSchema,
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
