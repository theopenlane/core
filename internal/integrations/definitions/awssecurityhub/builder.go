package awssecurityhub

import (
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// Builder returns the AWS Security Hub definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "aws",
				DisplayName: "AWS",
				Description: "Collect AWS Security Hub findings, AWS IAM users and groups, using a shared AWS assume-role credential.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws",
				Tags:        []string{"findings", "directory", "assets"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: providerkit.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         awsAssumeRoleCredential.ID(),
					Name:        "AWS Assume Role",
					Description: "Cross-account IAM role used to access Security Hub.",
					Schema:      awsAssumeRoleSchema,
					Recommended: true,
				},
				{
					Ref:         awsServiceAccountCredential.ID(),
					Name:        "AWS Static Credentials",
					Description: "Static IAM access keys for direct Security Hub access without assume-role.",
					Schema:      awsServiceAccountSchema,
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       awsAssumeRoleCredential.ID(),
					Name:                "AWS Assume Role",
					Description:         "Configure Security Hub access using a cross-account IAM role.",
					CredentialRefs:      []types.CredentialSlotID{awsAssumeRoleCredential.ID()},
					ClientRefs:          []types.ClientID{securityHubClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential.ID(),
						Description:   "Removes the stored IAM assume-role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.",
					},
				},
				{
					CredentialRef:       awsServiceAccountCredential.ID(),
					Name:                "AWS Static Credentials",
					Description:         "Configure Security Hub access using static IAM access keys.",
					CredentialRefs:      []types.CredentialSlotID{awsServiceAccountCredential.ID()},
					ClientRefs:          []types.ClientID{securityHubClient.ID()},
					ValidationOperation: healthCheckOperation.Name(),
					Integration:         installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsServiceAccountCredential.ID(),
						Description:   "Removes the stored IAM access key credentials from Openlane. If the IAM user is no longer needed, delete it from your AWS account.",
					},
				},
			},
			Clients: []types.ClientRegistration{
				{
					Ref:            securityHubClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsServiceAccountCredential.ID()},
					Description:    "AWS Security Hub client",
					Build:          SecurityHubClientBuilder{cfg: cfg}.Build,
				},
				{
					Ref:            configServiceClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsServiceAccountCredential.ID()},
					Description:    "AWS Config client",
					Build:          ConfigServiceClientBuilder{cfg: cfg}.Build,
				},
				{
					Ref:            iamClient.ID(),
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID(), awsServiceAccountCredential.ID()},
					Description:    "AWS IAM client",
					Build:          IAMClientBuilder{cfg: cfg}.Build,
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         healthCheckOperation.Name(),
					Description:  "Validate Security Hub access",
					Topic:        definitionID.OperationTopic(healthCheckOperation.Name()),
					ClientRef:    securityHubClient.ID(),
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle:       HealthCheck{}.Handle(),
					ConfigSchema: healthCheckSchema,
				},
				{
					Name:         findingsCollectOperation.Name(),
					Description:  "Collect AWS Security Hub for findings and vulnerability ingestion",
					Topic:        definitionID.OperationTopic(findingsCollectOperation.Name()),
					ClientRef:    securityHubClient.ID(),
					ConfigSchema: findingsCollectSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.FindingSync.Disable }),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaFinding,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaVulnerability,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"AWSSecurityHubReadOnlyAccess"},
				},
				{
					Name:         directorySyncOperation.Name(),
					Description:  "Sync AWS IAM users, groups, and memberships as directory accounts",
					Topic:        definitionID.OperationTopic(directorySyncOperation.Name()),
					ClientRef:    iamClient.ID(),
					ConfigSchema: directorySyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.DirectorySync.Disable }),
					Ingest: []types.IngestContract{
						{
							Schema:         integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
							ExhaustiveSync: true,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
						},
						{
							Schema: integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
						},
					},
					IngestHandle:        DirectorySync{}.IngestHandle(),
					RequiredPermissions: []string{"iam:ListUsers", "iam:ListGroups", "iam:ListGroupsForUser", "iam:ListUserTags"},
				},
				{
					Name:         checkSyncOperation.Name(),
					Description:  "Sync AWS Config rules and check results",
					Topic:        definitionID.OperationTopic(checkSyncOperation.Name()),
					ClientRef:    configServiceClient.ID(),
					ConfigSchema: checkSyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.CheckSync.Disable }),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaCheckResult,
						},
					},
					IngestHandle: CheckSync{}.IngestHandle(),
					RequiredPermissions: []string{
						"config:DescribeConfigRules",
						"config:DescribeComplianceByConfigRule",
						"controlcatalog:ListControls",
						"controlcatalog:ListControlMappings",
						"controlcatalog:ListCommonControls",
					},
					DisabledForAll: true,
				},
				{
					Name:         assetSyncOperation.Name(),
					Description:  "Sync assets from AWS",
					Topic:        definitionID.OperationTopic(assetSyncOperation.Name()),
					ClientRef:    securityHubClient.ID(),
					ConfigSchema: assetSyncSchema,
					Policy:       types.ExecutionPolicy{Reconcile: true},
					Disabled:     providerkit.DisabledWhen(func(u UserInput) bool { return u.AssetSync.Disable }),
					Ingest: []types.IngestContract{
						{
							Schema: integrationgenerated.IntegrationMappingSchemaAsset,
						},
					},
					IngestHandle:        AssetSync{}.IngestHandle(),
					RequiredPermissions: []string{"AWSSecurityHubReadOnlyAccess"},
					DisabledForAll:      true,
				},
			},
			Mappings: append(awsSecurityHubMappings(), awsIamMappings()...),
		}, nil
	})
}
