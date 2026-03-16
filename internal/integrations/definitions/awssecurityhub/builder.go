package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID                    = types.NewDefinitionRef("def_01K0AWSSECHUB0000000000001")
	HealthDefaultOperation          = types.NewOperationRef[struct{}]("health.default")
	VulnerabilitiesCollectOperation = types.NewOperationRef[struct{}]("vulnerabilities.collect")
)

const Slug = "aws_security_hub"

// userInput holds installation-specific configuration collected from the user
type userInput struct {
	Label         string   `json:"label,omitempty"         jsonschema:"title=Installation Label"`
	AccountScope  string   `json:"accountScope,omitempty"  jsonschema:"title=Account Scope"`
	AccountIDs    []string `json:"accountIds,omitempty"    jsonschema:"title=Account IDs"`
	LinkedRegions []string `json:"linkedRegions,omitempty" jsonschema:"title=Linked Regions"`
}

// credential holds the AWS STS role and optional static key material for one Security Hub installation
// Fields are named to match awskit.awsProviderData JSON tags so MetadataFromProviderData decodes them correctly
type credential struct {
	RoleARN         string   `json:"roleArn"                   jsonschema:"required,title=IAM Role ARN,description=Cross-account role Openlane should assume in the tenant environment."`
	ExternalID      string   `json:"externalId"                jsonschema:"required,title=External ID,description=External ID required in the tenant role trust policy."`
	HomeRegion      string   `json:"homeRegion"                jsonschema:"required,title=Security Hub Home Region,description=AWS region where Security Hub cross-region aggregation is managed (e.g. us-east-1)."`
	AccountScope    string   `json:"accountScope,omitempty"    jsonschema:"title=Account Scope,description=Collect from all delegated accounts or restrict to specific account IDs.,enum=all,enum=specific"`
	AccountIDs      []string `json:"accountIds,omitempty"      jsonschema:"title=Account IDs,description=Required when accountScope is specific."`
	LinkedRegions   []string `json:"linkedRegions,omitempty"   jsonschema:"title=Linked Regions,description=Filter findings to these source regions. Empty means all regions."`
	AccountID       string   `json:"accountId,omitempty"       jsonschema:"title=Account ID,description=AWS account ID for reference."`
	SessionName     string   `json:"sessionName,omitempty"     jsonschema:"title=Session Name,description=Optional STS session name override."`
	SessionDuration string   `json:"sessionDuration,omitempty" jsonschema:"title=Session Duration,description=Optional STS session duration (e.g. 1h)."`
	AccessKeyID     string   `json:"accessKeyId,omitempty"     jsonschema:"title=Access Key ID,description=Optional static source credential when runtime IAM is unavailable."`
	SecretAccessKey string   `json:"secretAccessKey,omitempty" jsonschema:"title=Secret Access Key"`
	SessionToken    string   `json:"sessionToken,omitempty"    jsonschema:"title=Session Token"`
}

// Builder returns the AWS Security Hub definition builder
func Builder() definition.Builder {
	return definition.Builder(func(_ context.Context) (types.Definition, error) {
		clientRef := types.NewClientRef[any]()

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
				Schema: providerkit.SchemaFrom[userInput](),
			},
			Credentials: &types.CredentialRegistration{
				Schema: providerkit.SchemaFrom[credential](),
			},
			Clients: []types.ClientRegistration{
				{
					Ref:         clientRef.ID(),
					Description: "AWS Security Hub client",
					Build:       buildSecurityHubClient,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:        HealthDefaultOperation.Name(),
					Description: "Validate Security Hub access via DescribeHub; confirms the assumed role can reach the hub in the configured home region",
					Topic:       HealthDefaultOperation.Topic(Slug),
					ClientRef:   clientRef.ID(),
					Policy:      types.ExecutionPolicy{Idempotent: true},
					Handle:      runHealthOperation,
				},
				{
					Name:         VulnerabilitiesCollectOperation.Name(),
					Description:  "Collect AWS Security Hub findings for vulnerability ingestion using server-side filters for severity, record state, and workflow status",
					Topic:        VulnerabilitiesCollectOperation.Topic(Slug),
					ClientRef:    clientRef.ID(),
					ConfigSchema: providerkit.SchemaFrom[findingsConfig](),
					Policy:       types.ExecutionPolicy{MaxRetries: 3, Idempotent: true},
					Ingest:       []types.IngestContract{{Schema: "vulnerability", EnsurePayloads: true}},
					Handle:       runVulnerabilitiesCollectOperation,
				},
			},
		}, nil
	})
}
