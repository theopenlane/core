package awssecurityhub

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// Builder returns the AWS Security Hub definition builder with the supplied operator config applied
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		installation := installationRef()

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID.ID(),
				Family:      "Amazon Web Services",
				DisplayName: "AWS",
				Description: "Collect AWS Security Hub findings, AWS IAM users and groups, using a shared AWS assume-role credential.",
				Category:    "security-posture",
				DocsURL:     "https://docs.theopenlane.io/docs/platform/integrations/aws",
				Tags:        []string{"findings", "directory", "assets"},
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: jsonx.SchemaFrom[UserInput](),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				awsAssumeRoleCredential.Registration(types.CredentialRegistration{
					Name:        "AWS Assume Role",
					Description: "Cross-account IAM role used to access Security Hub.",
					Schema:      awsAssumeRoleCredential.Schema(),
					Recommended: true,
				}),
				awsServiceAccountCredential.Registration(types.CredentialRegistration{
					Name:        "AWS Static Credentials",
					Description: "Static IAM access keys for direct Security Hub access without assume-role.",
					Schema:      awsServiceAccountCredential.Schema(),
				}),
			},
			Connections: []types.ConnectionRegistration{
				awsAssumeRoleConnection.Registration(types.ConnectionRegistration{
					Name:        "AWS Assume Role",
					Description: "Configure Security Hub access using a cross-account IAM role.",
					Meta: map[string]types.MetaInfo{
						"Openlane Principal ARN": {
							Value:     cfg.ARN,
							AllowCopy: true,
						},
					},
					CredentialRefs: []types.CredentialSlotID{awsAssumeRoleCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: securityHubClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsAssumeRoleCredential.ID(),
						Description:   "Removes the stored IAM assume-role configuration from Openlane. If the cross-account IAM role is no longer needed, delete it from your AWS account.",
					},
				}),
				awsServiceAccountConnection.Registration(types.ConnectionRegistration{
					Name:           "AWS Static Credentials",
					Description:    "Configure Security Hub access using static IAM access keys.",
					CredentialRefs: []types.CredentialSlotID{awsServiceAccountCredential.ID()},
					HealthCheck: &types.HealthCheckRegistration{
						ClientRef: securityHubClient.ID(),
						Handle:    HealthCheck{}.Handle(),
					},
					Integration: installation.Registration(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: awsServiceAccountCredential.ID(),
						Description:   "Removes the stored IAM access key credentials from Openlane. If the IAM user is no longer needed, delete it from your AWS account.",
					},
				}),
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
				findingsCollectOperation.Registration(definitionID, types.OperationRegistration{
					Description:    "Collect AWS Security Hub for findings and vulnerability ingestion",
					Policy:         types.ExecutionPolicy{Reconcile: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.FindingSync.Disable }),
					ConfigResolver: findingsCollectOperation.ConfigFrom(func(u UserInput) FindingSync { return u.FindingSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaFinding.Name,
						},
						{
							Schema: entityops.SchemaVulnerability.Name,
						},
					},
					IngestHandle:        FindingsCollect{}.IngestHandle(),
					RequiredPermissions: []string{"AWSSecurityHubReadOnlyAccess"},
				}),
				directorySyncOperation.Registration(definitionID, types.OperationRegistration{
					Description:    "Sync AWS IAM users, groups, and memberships as directory accounts",
					Policy:         types.ExecutionPolicy{Reconcile: true, Snapshot: true},
					Disabled:       providerkit.DisabledWhen(func(u UserInput) bool { return u.DirectorySync.Disable }),
					ConfigResolver: directorySyncOperation.ConfigFrom(func(u UserInput) DirectorySync { return u.DirectorySync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaDirectoryAccount.Name,
						},
						{
							Schema: entityops.SchemaDirectoryGroup.Name,
						},
						{
							Schema: entityops.SchemaDirectoryMembership.Name,
						},
					},
					IngestHandle:        DirectorySync{}.IngestHandle(),
					SkipDefaultLookback: true,
					RequiredPermissions: []string{"iam:ListUsers", "iam:ListGroups", "iam:ListGroupsForUser", "iam:ListUserTags"},
					Schedule:            gala.NewFullFetchSchedule(),
				}),
				checkSyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Sync AWS Config rules and check results",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					//  updated when DisabledForAll is removed
					Disabled:       providerkit.DisabledWhen(func(_ UserInput) bool { return true }),
					ConfigResolver: checkSyncOperation.ConfigFrom(func(u UserInput) CheckSync { return u.CheckSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaCheckResult.Name,
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
				}),
				assetSyncOperation.Registration(definitionID, types.OperationRegistration{
					Description: "Sync assets from AWS",
					Policy:      types.ExecutionPolicy{Reconcile: true},
					//  updated when DisabledForAll is removed
					Disabled:       providerkit.DisabledWhen(func(_ UserInput) bool { return true }),
					ConfigResolver: assetSyncOperation.ConfigFrom(func(u UserInput) AssetSync { return u.AssetSync }),
					Ingest: []types.IngestContract{
						{
							Schema: entityops.SchemaAsset.Name,
						},
					},
					IngestHandle:        AssetSync{}.IngestHandle(),
					RequiredPermissions: []string{"AWSSecurityHubReadOnlyAccess"},
					DisabledForAll:      true,
				}),
			},
			Mappings: []types.MappingRegistration{
				{
					Schema: entityops.SchemaFinding.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprFinding,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaControl.Name,
								TargetField:  control.FieldRefCode,
								SourceField:  entityops.FindingFields.Category.InputKey,
								SourceList:   entityops.FindingFields.Categories.InputKey,
							},
						},
					},
				},
				{
					Schema: entityops.SchemaVulnerability.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprVulnerability,
					},
				},
				{
					Schema: entityops.SchemaDirectoryAccount.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryAccount,
					},
				},
				{
					Schema: entityops.SchemaDirectoryGroup.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryGroup,
					},
				},
				{
					Schema: entityops.SchemaDirectoryMembership.Name,
					Spec: types.MappingOverride{
						FilterExpr: "true",
						MapExpr:    mapExprDirectoryMembership,
						Links: []types.LinkRule{
							{
								TargetSchema: entityops.SchemaDirectoryAccount.Name,
								TargetField:  directoryaccount.FieldExternalID,
								SourceField:  entityops.DirectoryMembershipFields.DirectoryAccountID.InputKey,
							},
							{
								TargetSchema: entityops.SchemaDirectoryGroup.Name,
								TargetField:  directorygroup.FieldExternalID,
								SourceField:  entityops.DirectoryMembershipFields.DirectoryGroupID.InputKey,
							},
						},
					},
				},
			},
		}, nil
	})
}
