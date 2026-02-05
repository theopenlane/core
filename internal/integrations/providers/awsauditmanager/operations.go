package awsauditmanager

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/common/integrations/helpers"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	awsAuditHealth types.OperationName = "health.default"

	awsAuditDefaultSession = "openlane-auditmanager"
)

// awsAuditOperations lists the AWS Audit Manager operations supported by this provider.
func awsAuditOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		helpers.HealthOperation(awsAuditHealth, "Validate AWS Audit Manager access by listing assessments.", ClientAWSAuditManager, runAWSAuditHealth),
	}
}

// runAWSAuditHealth validates AWS Audit Manager access via ListAssessments.
func runAWSAuditHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveAuditManagerClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	_, err = client.ListAssessments(ctx, &auditmanager.ListAssessmentsInput{
		MaxResults: aws.Int32(1),
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

type awsAuditManagerMetadata = helpers.AWSMetadata

// resolveAuditManagerClient returns a pooled client when available or builds one on demand.
func resolveAuditManagerClient(ctx context.Context, input types.OperationInput) (*auditmanager.Client, awsAuditManagerMetadata, error) {
	if client, ok := input.Client.(*auditmanager.Client); ok && client != nil {
		meta, err := awsAuditManagerMetadataFromPayload(input.Credential)
		if err != nil {
			return nil, awsAuditManagerMetadata{}, err
		}
		return client, meta, nil
	}

	return buildAuditManagerClient(ctx, input.Credential)
}

// buildAuditManagerClient constructs an Audit Manager client from the stored credential payload.
func buildAuditManagerClient(ctx context.Context, payload types.CredentialPayload) (*auditmanager.Client, awsAuditManagerMetadata, error) {
	meta, err := awsAuditManagerMetadataFromPayload(payload)
	if err != nil {
		return nil, awsAuditManagerMetadata{}, err
	}

	cfg, err := helpers.BuildAWSConfig(ctx, meta.Region, helpers.AWSCredentialsFromPayload(payload), helpers.AWSAssumeRole{
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

// awsAuditManagerMetadataFromPayload extracts AWS metadata required for Audit Manager.
func awsAuditManagerMetadataFromPayload(payload types.CredentialPayload) (awsAuditManagerMetadata, error) {
	meta := payload.Data.ProviderData
	if len(meta) == 0 {
		return awsAuditManagerMetadata{}, ErrMetadataMissing
	}

	parsed := helpers.AWSMetadataFromProviderData(meta, awsAuditDefaultSession)
	if err := helpers.RequireString(parsed.RoleARN, ErrRoleARNMissing); err != nil {
		return awsAuditManagerMetadata{}, err
	}
	if err := helpers.RequireString(parsed.Region, ErrRegionMissing); err != nil {
		return awsAuditManagerMetadata{}, err
	}

	return awsAuditManagerMetadata(parsed), nil
}
