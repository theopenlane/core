package awsassets

import (
	"context"
	"encoding/json"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const awsDefaultSessionName = "openlane-awsassets"

type awsHealthDetails struct {
	Region    string `json:"region,omitempty"`
	RoleARN   string `json:"roleArn,omitempty"`
	AccountID string `json:"accountId,omitempty"`
	ARN       string `json:"arn,omitempty"`
	UserID    string `json:"userId,omitempty"`
}

type awsAssetCollectionDetails struct {
	AccountID string `json:"accountId"`
	Region    string `json:"region"`
	Message   string `json:"message"`
}

// buildAWSClient builds the AWS client using STS AssumeRole
func buildAWSClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	meta, err := awskit.MetadataFromProviderData(req.Credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, fmt.Errorf("awsassets: metadata decode failed: %w", err)
	}

	if meta.RoleARN == "" {
		return nil, ErrRoleARNMissing
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.CredentialsFromMetadata(meta), awskit.AssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return nil, fmt.Errorf("awsassets: aws config build failed: %w", err)
	}

	return sts.NewFromConfig(cfg), nil
}

// runHealthOperation validates AWS credentials using STS GetCallerIdentity
func runHealthOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	stsClient, ok := client.(*sts.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, err
	}

	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("awsassets: STS GetCallerIdentity failed: %w", err)
	}

	details := awsHealthDetails{
		Region:  meta.Region,
		RoleARN: meta.RoleARN,
	}

	if accountID := awssdk.ToString(resp.Account); accountID != "" {
		details.AccountID = accountID
	}

	if arn := awssdk.ToString(resp.Arn); arn != "" {
		details.ARN = arn
	}

	if userID := awssdk.ToString(resp.UserId); userID != "" {
		details.UserID = userID
	}

	return jsonx.ToRawMessage(details)
}

// runAssetCollectionOperation collects AWS asset inventory
func runAssetCollectionOperation(ctx context.Context, _ *generated.Integration, credential types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	stsClient, ok := client.(*sts.Client)
	if !ok {
		return nil, ErrClientType
	}

	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, err
	}

	resp, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("awsassets: identity verification failed: %w", err)
	}

	return jsonx.ToRawMessage(awsAssetCollectionDetails{
		AccountID: awssdk.ToString(resp.Account),
		Region:    meta.Region,
		Message:   "aws asset collection ready",
	})
}
