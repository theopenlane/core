package awssecurityhub

import (
	"context"
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verifyAssumeRole derives the installation metadata from the role ARN, then verifies Security Hub access under the assumed role
func verifyAssumeRole(ctx context.Context, req types.ConnectionRequest[AssumeRoleCredentialSchema], c Client) (InstallationMetadata, error) {
	metadata, err := assumeRoleMetadata(req.Credential)
	if err != nil {
		return InstallationMetadata{}, err
	}

	if err := describeHub(ctx, c); err != nil {
		return InstallationMetadata{}, err
	}

	return metadata, nil
}

// assumeRoleMetadata derives the installation metadata from the assume-role credential, requiring the ARN account to agree with any configured account
func assumeRoleMetadata(cred AssumeRoleCredentialSchema) (InstallationMetadata, error) {
	account := arnAccountID(cred.RoleARN)

	switch {
	case account == "":
		return InstallationMetadata{}, fmt.Errorf("%w: %s", ErrRoleARNInvalid, cred.RoleARN)
	case cred.AccountID != "" && cred.AccountID != account:
		return InstallationMetadata{}, fmt.Errorf("%w: configured account %s, role arn account %s", ErrAccountIDMismatch, cred.AccountID, account)
	}

	return InstallationMetadata{
		RoleARN:       cred.RoleARN,
		HomeRegion:    cred.HomeRegion,
		AccountID:     account,
		AccountScope:  cred.AccountScope,
		AccountIDs:    cred.AccountIDs,
		LinkedRegions: cred.LinkedRegions,
	}, nil
}

// verifyStatic verifies Security Hub access under static credentials and derives the account identity via STS
func verifyStatic(ctx context.Context, req types.ConnectionRequest[ServiceAccountCredentialSchema], c Client) (InstallationMetadata, error) {
	if err := describeHub(ctx, c); err != nil {
		return InstallationMetadata{}, err
	}

	accountID, err := callerAccountID(ctx, c.Config)
	if err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{
		HomeRegion: req.Credential.Region,
		AccountID:  accountID,
	}, nil
}

// describeHub validates Security Hub access by calling DescribeHub
func describeHub(ctx context.Context, c Client) error {
	if _, err := c.SecurityHub().DescribeHub(ctx, &securityhub.DescribeHubInput{}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error describing hub")

		if strings.Contains(err.Error(), "not subscribed to AWS Security Hub") {
			return ErrSecurityHubNotEnabled
		}

		return ErrDescribeHubFailed
	}

	return nil
}

// callerAccountID returns the AWS account id via STS GetCallerIdentity
func callerAccountID(ctx context.Context, cfg awssdk.Config) (string, error) {
	identity, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCallerIdentityFetchFailed, err)
	}

	accountID := awssdk.ToString(identity.Account)
	if accountID == "" {
		return "", ErrAccountIDMissing
	}

	return accountID, nil
}
