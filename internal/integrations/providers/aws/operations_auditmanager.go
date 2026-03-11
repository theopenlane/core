package aws

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	awsAuditAssessmentsOp types.OperationName = "audit_manager.assessments.list"
	awsAuditListMaxOne    int32               = 1
)

type awsAuditManagerDetails struct {
	RoleArn   string `json:"roleArn,omitempty"`
	Region    string `json:"region,omitempty"`
	AccountID string `json:"accountId,omitempty"`
}

// awsAuditManagerOperations lists the AWS Audit Manager operations supported by this provider
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

// runAWSAuditAssessments validates AWS Audit Manager access via ListAssessments
func runAWSAuditAssessments(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, meta, err := resolveAuditManagerClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	_, err = client.ListAssessments(ctx, &auditmanager.ListAssessmentsInput{
		MaxResults: awssdk.Int32(awsAuditListMaxOne),
	})
	if err != nil {
		return providerkit.OperationFailure("AWS Audit Manager list assessments failed", err, awsAuditManagerDetails{
			Region: meta.Region,
		})
	}

	details := awsAuditManagerDetails{
		RoleArn: meta.RoleARN,
		Region:  meta.Region,
	}

	if meta.AccountID != "" {
		details.AccountID = meta.AccountID
	}

	summary := "AWS Audit Manager reachable"
	if meta.AccountID != "" {
		summary = fmt.Sprintf("AWS Audit Manager reachable for account %s", meta.AccountID)
	}

	return providerkit.OperationSuccess(summary, details), nil
}
