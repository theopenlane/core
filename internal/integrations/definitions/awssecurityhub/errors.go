package awssecurityhub

import "errors"

var (
	// ErrRoleARNMissing indicates the IAM role ARN is missing from the credential
	ErrRoleARNMissing = errors.New("awssecurityhub: roleArn required")
	// ErrRoleARNInvalid is returned when the assume-role ARN does not carry an account id
	ErrRoleARNInvalid = errors.New("awssecurityhub: assume role arn missing account id")
	// ErrRegionMissing indicates the home region is missing from the credential
	ErrRegionMissing = errors.New("awssecurityhub: homeRegion required")
	// ErrAWSConfigBuildFailed indicates the AWS SDK config could not be constructed
	ErrAWSConfigBuildFailed = errors.New("awssecurityhub: aws config build failed")
	// ErrCallerIdentityFetchFailed indicates STS GetCallerIdentity failed
	ErrCallerIdentityFetchFailed = errors.New("awssecurityhub: caller identity fetch failed")
	// ErrAccountIDMissing indicates STS GetCallerIdentity returned no account id
	ErrAccountIDMissing = errors.New("awssecurityhub: account id missing")
	// ErrAccountIDMismatch indicates the configured account id disagrees with the role ARN
	ErrAccountIDMismatch = errors.New("awssecurityhub: configured account id does not match role arn account")
	// ErrDescribeHubFailed indicates DescribeHub failed
	ErrDescribeHubFailed = errors.New("awssecurityhub: describe hub failed")
	// ErrSecurityHubNotEnabled indicates security hub is not enabled for the account
	ErrSecurityHubNotEnabled = errors.New("awssecurityhub: security hub not enabled for account")
	// ErrFindingsFetchFailed indicates GetFindings failed
	ErrFindingsFetchFailed = errors.New("awssecurityhub: findings fetch failed")
	// ErrFindingEncode indicates a finding payload could not be serialized
	ErrFindingEncode = errors.New("awssecurityhub: finding encode failed")
	// ErrIAMUsersFetchFailed indicates ListUsers failed
	ErrIAMUsersFetchFailed = errors.New("awsiam: IAM users fetch failed")
	// ErrIAMGroupsFetchFailed indicates ListGroups failed
	ErrIAMGroupsFetchFailed = errors.New("awsiam: IAM groups fetch failed")
	// ErrIAMGroupsForUserFetchFailed indicates ListGroupsForUser failed
	ErrIAMGroupsForUserFetchFailed = errors.New("awsiam: IAM groups for user fetch failed")
	// ErrDirectorySyncPayloadEncode indicates a directory sync payload could not be serialized
	ErrDirectorySyncPayloadEncode = errors.New("awsiam: directory sync payload encode failed")
)
