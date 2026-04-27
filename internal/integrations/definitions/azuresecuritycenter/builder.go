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
				Family:      "Azure",
				DisplayName: "Microsoft Defender for Cloud",
				Description: "Collect security assessment findings and vulnerability data from Microsoft Defender for Cloud across an Azure subscription.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/azure_security_center/overview",
				Tags:        []string{"vulnerabilities", "assets"},
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
					Description: "Azure service principal used to access Microsoft Defender for Cloud.",
					Schema:      securityCenterSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       securityCenterCredential.ID(),
					Name:                "Azure Service Principal",
					Description:         "Configure Defender for Cloud access using an Azure service principal.",
					CredentialRefs:      []types.CredentialSlotID{securityCenterCredential.ID()},
					ClientRefs:          []types.ClientID{securityCenterClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: securityCenterCredential.ID(),
						Description:   "Removes the stored service principal credentials from Openlane. If the Azure app registration is no longer needed, delete it from your Azure tenant.",
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
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    securityCenterClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         assessmentsCollectOperation.Name(),
					Description:  "Collect unhealthy security posture assessment findings for vulnerability ingestion",
					Topic:        definitionID.OperationTopic(assessmentsCollectOperation.Name()),
					ClientRef:    securityCenterClient.ID(),
					ConfigSchema: assessmentsCollectSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
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
					Topic:        definitionID.OperationTopic(subAssessmentsCollectOperation.Name()),
					ClientRef:    securityCenterClient.ID(),
					ConfigSchema: subAssessmentsCollectSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
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
