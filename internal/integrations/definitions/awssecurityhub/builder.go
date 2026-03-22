package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the AWS Security Hub definition builder
func Builder() definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Slug:        Slug,
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
					Ref:         awsAssumeRoleCredential,
					Name:        "AWS Assume Role",
					Description: "Assume-role and collection-scope credential slot shared by the AWS service clients in this definition.",
					Schema:      providerkit.SchemaFrom[AssumeRoleCredentialSchema](),
				},
				{
					Ref:         awsSourceCredential,
					Name:        "AWS Source Credential",
					Description: "Optional static source credential used to assume the configured AWS role when runtime IAM is unavailable.",
					Schema:      providerkit.SchemaFrom[SourceCredentialSchema](),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       awsAssumeRoleCredential,
					Name:                "AWS Assume Role",
					Description:         "Configure AWS Security Hub and Audit Manager access using a cross-account IAM role and optional source credentials for STS.",
					CredentialRefs:      []types.CredentialRef{awsAssumeRoleCredential, awsSourceCredential},
					ClientRefs:          []types.ClientID{SecurityHubClient.ID(), AuditManagerClient.ID()},
					ValidationOperation: HealthDefaultOperation.Name(),
					Installation:        Installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential,
						Name:          "Disconnect AWS Assume Role",
						Description:   "Remove the persisted AWS assume-role configuration and optional source credential used by this installation.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            SecurityHubClient.ID(),
					CredentialRefs: []types.CredentialRef{awsAssumeRoleCredential, awsSourceCredential},
					Description:    "AWS Security Hub client",
					Build:          SecurityHubClientBuilder{}.Build,
				},
				{
					Ref:            AuditManagerClient.ID(),
					CredentialRefs: []types.CredentialRef{awsAssumeRoleCredential, awsSourceCredential},
					Description:    "AWS Audit Manager client",
					Build:          AuditManagerClientBuilder{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate Security Hub access",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SecurityHubClient.ID(),
					Handle:      HealthCheck{}.Handle(),
				},
				{
					Name:         AssessmentsListOperation.Name(),
					Description:  "List AWS Audit Manager assessments for review and future ingest support",
					Topic:        AssessmentsListOperation.Topic(Slug),
					ClientRef:    AuditManagerClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[AssessmentsConfig](),
					Handle:       AssessmentsList{}.Handle(),
				},
				{
					Name:         VulnerabilitiesCollectOperation.Name(),
					Description:  "Collect AWS Security Hub findings for vulnerability ingestion",
					Topic:        VulnerabilitiesCollectOperation.Topic(Slug),
					ClientRef:    SecurityHubClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[FindingsConfig](),
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
