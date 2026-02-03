package awsauditmanager

import (
	"context"

	"github.com/theopenlane/core/common/integrations/types"
)

const (
	// ClientAWSAuditManager identifies the AWS Audit Manager client descriptor.
	ClientAWSAuditManager types.ClientName = "auditmanager"
)

// awsAuditManagerClientDescriptors returns the AWS Audit Manager client descriptors for pooling.
func awsAuditManagerClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:     TypeAWSAuditManager,
			Name:         ClientAWSAuditManager,
			Description:  "AWS Audit Manager client",
			Build:        buildAWSAuditManagerClient,
			ConfigSchema: map[string]any{"type": "object"},
		},
	}
}

// buildAWSAuditManagerClient builds the AWS Audit Manager client for pooling.
func buildAWSAuditManagerClient(ctx context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	client, _, err := buildAuditManagerClient(ctx, payload)
	if err != nil {
		return nil, err
	}
	return client, nil
}
