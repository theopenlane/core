package awsauditmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/theopenlane/core/internal/integrations/types"
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

func runAWSAuditHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	_ = ctx
	meta := input.Credential.Data.ProviderData
	if len(meta) == 0 {
		return types.OperationResult{}, fmt.Errorf("aws audit manager: provider metadata missing")
	}

	account := strings.TrimSpace(stringValue(meta, "accountId"))
	region := strings.TrimSpace(stringValue(meta, "region"))
	roleArn := strings.TrimSpace(stringValue(meta, "roleArn"))

	if roleArn == "" {
		return types.OperationResult{}, fmt.Errorf("aws audit manager: roleArn missing")
	}
	if region == "" {
		return types.OperationResult{}, fmt.Errorf("aws audit manager: region missing")
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

func stringValue(data map[string]any, key string) string {
	if len(data) == 0 {
		return ""
	}
	value, ok := data[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}
