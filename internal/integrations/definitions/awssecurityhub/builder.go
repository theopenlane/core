package awssecurityhub

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// Builder returns the AWS Security Hub definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			ID:           definitionID.ID(),
			Family:       "Amazon Web Services",
			DisplayName:  "AWS",
			Description:  "Collect AWS Security Hub findings, AWS IAM users and groups, using a shared AWS assume-role credential.",
			Category:     "security-posture",
			DocsURL:      "https://docs.theopenlane.io/docs/platform/integrations/aws",
			Tags:         []string{"findings", "directory", "assets"},
			Active:       true,
			Visible:      true,
			Installation: installation.Registration(),
			Connections: []types.Connector{
				assumeRole.
					Name("AWS Assume Role").
					Description("Configure Security Hub access using a cross-account IAM role.").
					Recommended().
					Meta(map[string]types.MetaInfo{
						"Openlane Principal ARN": {
							Value:     cfg.ARN,
							AllowCopy: true,
						},
					}).
					Provides(assumeRoleClient(cfg)).
					Verified(verifyAssumeRole).
					Disconnects("Removes the stored IAM assume-role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.", nil),
				staticCredentials.
					Name("AWS Static Credentials").
					Description("Configure Security Hub access using static IAM access keys.").
					Provides(staticClient).
					Verified(verifyStatic).
					Disconnects("Removes the stored IAM access key credentials from Openlane. If the IAM user is no longer needed, delete it from your AWS account.", nil),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[FindingSync]().
					Ingests(runFindingsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(
						types.IngestContract{Schema: entityops.SchemaFinding.Name},
						types.IngestContract{Schema: entityops.SchemaVulnerability.Name},
					).
					Permissions("AWSSecurityHubReadOnlyAccess").
					Description("Collect AWS Security Hub for findings and vulnerability ingestion").
					Registration(),
				types.OperationRefOf[providerkit.DirectorySync]().
					Ingests(runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("iam:ListUsers", "iam:ListGroups", "iam:ListGroupsForUser", "iam:ListUserTags").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Description("Sync AWS IAM users, groups, and memberships as directory accounts").
					Registration(),
				types.OperationRefOf[CheckSync]().
					Ingests(runCheckSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaCheckResult.Name}).
					Permissions(
						"config:DescribeConfigRules",
						"config:DescribeComplianceByConfigRule",
						"controlcatalog:ListControls",
						"controlcatalog:ListControlMappings",
						"controlcatalog:ListCommonControls",
					).
					DisabledForAll(true).
					Description("Sync AWS Config rules and check results").
					Registration(),
				types.OperationRefOf[AssetSync]().
					Ingests(runAssetSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("AWSSecurityHubReadOnlyAccess").
					DisabledForAll(true).
					Description("Sync assets from AWS").
					Registration(),
			},
			Mappings: append([]types.MappingRegistration{
				providerkit.FindingMapping(mapExprFinding),
				{
					Schema: entityops.SchemaVulnerability.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprVulnerability,
					},
				},
			}, providerkit.DirectoryMappings(mapExprDirectoryAccount, mapExprDirectoryGroup, mapExprDirectoryMembership)...),
		}, nil
	})
}
