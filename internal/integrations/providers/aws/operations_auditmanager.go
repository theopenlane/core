package aws

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	awsAuditAssessmentsOp types.OperationName = "audit_manager.assessments.list"
)

// awsAuditManagerOperations lists the AWS Audit Manager operations supported by this provider.
func awsAuditManagerOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        awsAuditAssessmentsOp,
			Kind:        types.OperationKindScanSettings,
			Description: "List Audit Manager assessments to validate access.",
			Client:      ClientAWSAuditManager,
			Run:         runAWSAuditAssessments,
		},
	}
}

// runAWSAuditAssessments validates AWS Audit Manager access via ListAssessments.
func runAWSAuditAssessments(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveAuditManagerClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	_, err = client.ListAssessments(ctx, &auditmanager.ListAssessmentsInput{
		MaxResults: awssdk.Int32(1),
	})
	if err != nil {
		return types.OperationResult{
			Status:  types.OperationStatusFailed,
			Summary: "AWS Audit Manager list assessments failed",
			Details: map[string]any{
				"region": meta.Region,
				"error":  err.Error(),
			},
		}, err
	}

	details := map[string]any{
		"roleArn": meta.RoleARN,
		"region":  meta.Region,
	}
	if meta.AccountID != "" {
		details["accountId"] = meta.AccountID
	}

	summary := "AWS Audit Manager reachable"
	if meta.AccountID != "" {
		summary = fmt.Sprintf("AWS Audit Manager reachable for account %s", meta.AccountID)
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: details,
	}, nil
}

// resolveAuditManagerClient returns a pooled client when available or builds one on demand.
func resolveAuditManagerClient(ctx context.Context, input types.OperationInput) (*auditmanager.Client, auth.AWSMetadata, error) {
	if client, ok := input.Client.(*auditmanager.Client); ok && client != nil {
		meta, err := awsMetadataFromPayload(input.Credential, awsDefaultSession)
		if err != nil {
			return nil, auth.AWSMetadata{}, err
		}
		return client, meta, nil
	}

	return buildAuditManagerClient(ctx, input.Credential)
}

// buildAuditManagerClient constructs an Audit Manager client from the stored credential payload.
func buildAuditManagerClient(ctx context.Context, payload types.CredentialPayload) (*auditmanager.Client, auth.AWSMetadata, error) {
	meta, err := awsMetadataFromPayload(payload, awsDefaultSession)
	if err != nil {
		return nil, auth.AWSMetadata{}, err
	}

	cfg, err := auth.BuildAWSConfig(ctx, meta.Region, auth.AWSCredentialsFromPayload(payload), auth.AWSAssumeRole{
		RoleARN:         meta.RoleARN,
		ExternalID:      meta.ExternalID,
		SessionName:     meta.SessionName,
		SessionDuration: meta.SessionDuration,
	})
	if err != nil {
		return nil, meta, err
	}

	return auditmanager.NewFromConfig(cfg), meta, nil
}
