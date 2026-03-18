package aws

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const awsHealthDefault types.OperationName = "health.default"

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

	cfg, err := auth.BuildAWSConfig(ctx, meta.Region, auth.AWSCredentialsFromPayload(input.Credential), auth.AWSAssumeRole{
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
	details := map[string]any{
		"region":  meta.Region,
		"roleArn": meta.RoleARN,
	}
	if accountID != "" {
		details["accountId"] = accountID
	}
	if arn := awssdk.ToString(resp.Arn); arn != "" {
		details["arn"] = arn
	}
	if userID := awssdk.ToString(resp.UserId); userID != "" {
		details["userId"] = userID
	}

	summary := "AWS credentials verified"
	if accountID != "" {
		summary = fmt.Sprintf("AWS credentials verified for account %s", accountID)
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: details,
	}, nil
}
