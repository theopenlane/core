package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the AWS Security Hub definition builder
func Builder() registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "aws",
				DisplayName: "AWS Security Hub",
				Description: "Collect AWS Security Hub findings and Audit Manager summaries using a shared AWS assume-role credential.",
				Category:    "cloud",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
				Labels:      map[string]string{"vendor": "aws"},
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
					Description: "Assume-role and collection-scope credential slot shared by the AWS service clients in this definition.",
					Schema:      awsAssumeRoleSchema,
				},
				{
					Ref:         awsSourceCredential.ID(),
					Name:        "AWS Source Credential",
					Description: "Optional static source credential used to assume the configured AWS role when runtime IAM is unavailable.",
					Schema:      awsSourceSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       awsAssumeRoleCredential.ID(),
					Name:                "AWS Assume Role",
					Description:         "Configure AWS Security Hub and Audit Manager access using a cross-account IAM role and optional source credentials for STS.",
					CredentialRefs:      []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsSourceCredential.ID()},
					ClientRefs:          []types.ClientID{SecurityHubClient.ID(), AuditManagerClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential.ID(),
						Name:          "Disconnect AWS Assume Role",
						Description:   "Remove the persisted AWS assume-role configuration and optional source credential used by this installation.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            SecurityHubClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsSourceCredential.ID()},
					Description:    "AWS Security Hub client",
					Build:          SecurityHubClientBuilder{}.Build,
				},
				{
					Ref:            AuditManagerClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsSourceCredential.ID()},
					Description:    "AWS Audit Manager client",
					Build:          AuditManagerClientBuilder{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate Security Hub access",
					Topic:       types.OperationTopic(DefinitionID.ID(), HealthDefaultOperation.Name()),
					ClientRef:   SecurityHubClient.ID(),
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:         AssessmentsCollectOperation.Name(),
					Description:  "Collect AWS Audit Manager assessments as findings",
					Topic:        types.OperationTopic(DefinitionID.ID(), AssessmentsCollectOperation.Name()),
					ClientRef:    AuditManagerClient.ID(),
					ConfigSchema: assessmentsCollectSchema,
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaFinding,
						},
					},
					IngestHandle: AssessmentsCollect{}.IngestHandle(),
				},
				{
					Name:         VulnerabilitiesCollectOperation.Name(),
					Description:  "Collect AWS Security Hub findings for vulnerability ingestion",
					Topic:        types.OperationTopic(DefinitionID.ID(), VulnerabilitiesCollectOperation.Name()),
					ClientRef:    SecurityHubClient.ID(),
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
