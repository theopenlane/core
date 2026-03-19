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
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[CredentialSchema](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         SecurityHubClient.ID(),
					Description: "AWS Security Hub client",
					Build:       SecurityHubClientBuilder{}.Build,
				},
				{
					Ref:         AuditManagerClient.ID(),
					Description: "AWS Audit Manager client",
					Build:       AuditManagerClientBuilder{}.Build,
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
