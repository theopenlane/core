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
				Version:     "v1",
				Family:      "aws",
				DisplayName: "AWS Security Hub",
				Description: "Collect AWS Security Hub findings for vulnerability ingestion using STS role assumption in a tenant AWS environment.",
				Category:    "cloud",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws/overview",
				Labels:      map[string]string{"vendor": "aws", "service": "security-hub"},
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
					Build:       Client{}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate Security Hub access via DescribeHub; confirms the assumed role can reach the hub in the configured home region",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   SecurityHubClient.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      HealthCheck{}.Handle(Client{}),
				},
				{
					Name:         VulnerabilitiesCollectOperation.Name(),
					Description:  "Collect AWS Security Hub findings for vulnerability ingestion using server-side filters for severity, record state, and workflow status",
					Topic:        VulnerabilitiesCollectOperation.Topic(Slug),
					ClientRef:    SecurityHubClient.ID(),
					ConfigSchema: providerkit.SchemaFrom[FindingsConfig](),
					Policy:       types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaVulnerability,
							EnsurePayloads: true,
						},
					},
					Handle: VulnerabilitiesCollect{}.Handle(Client{}),
				},
			},
			Mappings: awsSecurityHubMappings(),
		}, nil
	})
}
