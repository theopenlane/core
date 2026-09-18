package awssecurityhub

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// installationRef builds the typed installation metadata handle for the AWS Security Hub definition
func installationRef() types.InstallationRef[InstallationMetadata] {
	return types.NewInstallationRef(resolveInstallationMetadata)
}

// resolveInstallationMetadata derives AWS connection metadata from whichever credential is bound.
// It uses the assume-role credential when present, otherwise falls back to the service account credential.
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	_, hasAssumeRole := req.Credentials.Resolve(awsAssumeRoleCredential.ID())
	if !hasAssumeRole {
		return resolveServiceAccountInstallationMetadata(ctx, req)
	}

	assumeRole, ok, err := awsAssumeRoleCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error resolving assume role credentials")
		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
	}

	if !ok {
		return InstallationMetadata{}, ok, nil
	}

	account := arnAccountID(assumeRole.RoleARN)
	if account == "" {
		return InstallationMetadata{}, false, fmt.Errorf("%w: %s", ErrRoleARNInvalid, assumeRole.RoleARN)
	}

	if assumeRole.AccountID != "" && assumeRole.AccountID != account {
		return InstallationMetadata{}, false, fmt.Errorf("%w: configured account %s, role arn account %s", ErrAccountIDMismatch, assumeRole.AccountID, account)
	}

	metadata := InstallationMetadata{
		RoleARN:       assumeRole.RoleARN,
		HomeRegion:    assumeRole.HomeRegion,
		AccountID:     account,
		AccountScope:  assumeRole.AccountScope,
		AccountIDs:    assumeRole.AccountIDs,
		LinkedRegions: assumeRole.LinkedRegions,
	}

	return metadata, true, nil
}

// resolveServiceAccountInstallationMetadata derives the AWS account identity for the static credential path from STS
func resolveServiceAccountInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	serviceAccount, ok, err := awsServiceAccountCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("awssecurityhub: error resolving aws credentials")
		return InstallationMetadata{}, false, ErrCredentialMetadataInvalid
	}

	if !ok {
		return InstallationMetadata{}, false, nil
	}

	cfg, err := buildAWSConfigFromStaticCreds(ctx, serviceAccount)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	accountID, err := callerAccountID(ctx, cfg)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	return InstallationMetadata{
		HomeRegion: serviceAccount.Region,
		AccountID:  accountID,
	}, true, nil
}

// callerAccountID returns the AWS account id that owns the credentials in the config via STS GetCallerIdentity
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
