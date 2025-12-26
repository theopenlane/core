package awsauditmanager

import (
	"context"
	"fmt"

	"github.com/theopenlane/common/integrations/helpers"
	"github.com/theopenlane/common/integrations/types"
)

const (
	awsAuditHealth types.OperationName = "health.default"
)

func awsAuditOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		{
			Name:        awsAuditHealth,
			Kind:        types.OperationKindHealth,
			Description: "Validate stored AWS Audit Manager credentials.",
			Run:         runAWSAuditHealth,
		},
	}
}

// runAWSAuditHealth validates that the AWS Audit Manager credentials are present
func runAWSAuditHealth(_ context.Context, input types.OperationInput) (types.OperationResult, error) {
	meta := input.Credential.Data.ProviderData
	if len(meta) == 0 {
		return types.OperationResult{}, ErrMetadataMissing
	}

	account := helpers.StringValue(meta, "accountId")
	region := helpers.StringValue(meta, "region")
	roleArn := helpers.StringValue(meta, "roleArn")

	if roleArn == "" {
		return types.OperationResult{}, ErrRoleARNMissing
	}
	if region == "" {
		return types.OperationResult{}, ErrRegionMissing
	}

	details := map[string]any{
		"roleArn": roleArn,
		"region":  region,
	}
	if account != "" {
		details["accountId"] = account
	}

	summary := "AWS Audit Manager credential present"
	if account != "" {
		summary = fmt.Sprintf("AWS Audit Manager credential present for account %s", account)
	}

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: summary,
		Details: details,
	}, nil
}
