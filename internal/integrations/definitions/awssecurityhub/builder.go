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
			HealthCheck:  securityHubClient.HealthCheck(checkHealth),
			Installation: installation.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				awsAssumeRoleCredential.Registration(types.CredentialRegistration{
					Name:        "AWS Assume Role",
					Description: "Cross-account IAM role used to access Security Hub.",
					Recommended: true,
				}),
				awsServiceAccountCredential.Registration(types.CredentialRegistration{
					Name:        "AWS Static Credentials",
					Description: "Static IAM access keys for direct Security Hub access without assume-role.",
				}),
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef: awsAssumeRoleCredential.ID(),
					Name:          "AWS Assume Role",
					Description:   "Configure Security Hub access using a cross-account IAM role.",
					Meta: map[string]types.MetaInfo{
						"Openlane Principal ARN": {
							Value:     cfg.ARN,
							AllowCopy: true,
						},
					},
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential.ID(),
						Description:   "Removes the stored IAM assume-role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.",
					},
				},
				{
					CredentialRef: awsServiceAccountCredential.ID(),
					Name:          "AWS Static Credentials",
					Description:   "Configure Security Hub access using static IAM access keys.",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsServiceAccountCredential.ID(),
						Description:   "Removes the stored IAM access key credentials from Openlane. If the IAM user is no longer needed, delete it from your AWS account.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				securityHubClient.Registration(SecurityHubClientBuilder{cfg: cfg}.Build, types.ClientRegistration{
					Description: "AWS Security Hub client",
				}),
				configServiceClient.Registration(ConfigServiceClientBuilder{cfg: cfg}.Build, types.ClientRegistration{
					Description: "AWS Config client",
				}),
				iamClient.Registration(IAMClientBuilder{cfg: cfg}.Build, types.ClientRegistration{
					Description: "AWS IAM client",
				}),
			},
			Operations: []types.OperationRegistration{
				types.OperationRefOf[FindingSync]().
					Ingests(securityHubClient, runFindingsCollect).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(
						types.IngestContract{Schema: entityops.SchemaFinding.Name},
						types.IngestContract{Schema: entityops.SchemaVulnerability.Name},
					).
					Permissions("AWSSecurityHubReadOnlyAccess").
					Registration(definitionID, types.OperationRegistration{
						Description: "Collect AWS Security Hub for findings and vulnerability ingestion",
					}),
				types.OperationRefOf[DirectorySync]().
					Ingests(iamClient, runDirectorySync).
					Policy(types.ExecutionPolicy{Reconcile: true, Snapshot: true}).
					Ingest(providerkit.DirectoryIngestContracts()...).
					Permissions("iam:ListUsers", "iam:ListGroups", "iam:ListGroupsForUser", "iam:ListUserTags").
					Schedule(gala.NewFullFetchSchedule()).
					SkipDefaultLookback().
					Registration(definitionID, types.OperationRegistration{
						Description: "Sync AWS IAM users, groups, and memberships as directory accounts",
					}),
				types.OperationRefOf[CheckSync]().
					Ingests(configServiceClient, runCheckSync).
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
					Registration(definitionID, types.OperationRegistration{
						Description: "Sync AWS Config rules and check results",
					}),
				types.OperationRefOf[AssetSync]().
					Ingests(configServiceClient, runAssetSync).
					Policy(types.ExecutionPolicy{Reconcile: true}).
					Ingest(types.IngestContract{Schema: entityops.SchemaAsset.Name}).
					Permissions("AWSSecurityHubReadOnlyAccess").
					DisabledForAll(true).
					Registration(definitionID, types.OperationRegistration{
						Description: "Sync assets from AWS",
					}),
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
