package aws

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/integrations/operations"
	awskit "github.com/theopenlane/core/internal/integrations/providers/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const awsHealthDefault types.OperationName = types.OperationHealthDefault

type awsHealthDetails struct {
	Region    string `json:"region,omitempty"`
	RoleArn   string `json:"roleArn,omitempty"`
	AccountID string `json:"accountId,omitempty"`
	ARN       string `json:"arn,omitempty"`
	UserID    string `json:"userId,omitempty"`
}

// awsHealthOperation builds the AWS health operation descriptor
func awsHealthOperation() types.OperationDescriptor {
	return operations.HealthOperation(awsHealthDefault, "Validate AWS access via STS GetCallerIdentity.", "", runAWSHealth)
}

// runAWSHealth validates credentials using STS GetCallerIdentity.
func runAWSHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta, err := awsMetadataFromPayload(input.Credential, awsDefaultSession)
	if err != nil {
		return types.OperationResult{}, err
	}

	cfg, err := awskit.BuildAWSConfig(ctx, meta.Region, awskit.AWSCredentialsFromPayload(input.Credential), awskit.AWSAssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return operations.OperationFailure("AWS config build failed", err, nil)
	}

	client := sts.NewFromConfig(cfg)
	resp, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return operations.OperationFailure("AWS STS GetCallerIdentity failed", err, nil)
	}

	accountID := awssdk.ToString(resp.Account)
	details := awsHealthDetails{
		Region:  meta.Region,
		RoleArn: meta.RoleARN,
	}
	if accountID != "" {
		details.AccountID = accountID
	}
	if arn := awssdk.ToString(resp.Arn); arn != "" {
		details.ARN = arn
	}
	if userID := awssdk.ToString(resp.UserId); userID != "" {
		details.UserID = userID
	}

	summary := "AWS credentials verified"
	if accountID != "" {
		summary = fmt.Sprintf("AWS credentials verified for account %s", accountID)
	}

	return operations.OperationSuccess(summary, details), nil
}
