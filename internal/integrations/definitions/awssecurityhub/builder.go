package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the AWS Security Hub definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "aws",
				DisplayName: "AWS Security Hub",
				Description: "Collect AWS Security Hub findings and Audit Manager summaries using a shared AWS assume-role credential.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
				Tags:        []string{"vulnerabilities", "assets"},
				Active:      true,
				Visible:     false,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         awsAssumeRoleCredential.ID(),
					Name:        "AWS Assume Role",
					Description: "Cross-account IAM role used to access Security Hub and Audit Manager.",
					Schema:      awsAssumeRoleSchema,
				},
				{
					Ref:         awsServiceAccountCredential.ID(),
					Name:        "AWS Source Credential",
					Description: "Optional IAM credentials used to assume the cross-account role when runtime IAM is unavailable.",
					Schema:      awsServiceAccountSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       awsAssumeRoleCredential.ID(),
					Name:                "AWS Assume Role",
					Description:         "Configure Security Hub and Audit Manager access using a cross-account IAM role.",
					CredentialRefs:      []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsServiceAccountCredential.ID()},
					ClientRefs:          []types.ClientID{securityHubClient.ID(), auditManagerClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Installation:        installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential.ID(),
						Description:   "Removes the stored IAM assume-role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            securityHubClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsServiceAccountCredential.ID()},
					Description:    "AWS Security Hub client",
					Build:          SecurityHubClientBuilder{}.Build,
				},
				{
					Ref:            auditManagerClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsServiceAccountCredential.ID()},
					Description:    "AWS Audit Manager client",
					Build:          AuditManagerClientBuilder{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Validate Security Hub access",
					Topic:        types.OperationTopic(definitionID.ID(), healthCheckOperation.Name()),
					ClientRef:    securityHubClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         assessmentsCollectOperation.Name(),
					Description:  "Collect AWS Audit Manager assessments as findings",
					Topic:        types.OperationTopic(definitionID.ID(), assessmentsCollectOperation.Name()),
					ClientRef:    auditManagerClient.ID(),
					ConfigSchema: assessmentsCollectSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaFinding,
						},
					},
					IngestHandle: AssessmentsCollect{}.IngestHandle(),
				},
				{
					Name:         vulnerabilitiesCollectOperation.Name(),
					Description:  "Collect AWS Security Hub findings for vulnerability ingestion",
					Topic:        types.OperationTopic(definitionID.ID(), vulnerabilitiesCollectOperation.Name()),
					ClientRef:    securityHubClient.ID(),
					ConfigSchema: vulnerabilitiesCollectSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle: VulnerabilitiesCollect{}.IngestHandle(),
				},
			},
			Mappings: awsSecurityHubMappings(),
		}, nil
	})
}
