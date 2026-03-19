package awsauditmanager

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an AWS Audit Manager health check
type HealthCheck struct {
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// RoleARN is the assumed role ARN when present
	RoleARN string `json:"roleArn,omitempty"`
	// AccountID is the AWS account identifier
	AccountID string `json:"accountId,omitempty"`
	// AccountStatus is the Audit Manager account status
	AccountStatus string `json:"accountStatus"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := AuditManagerClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

// Run validates Audit Manager access via GetAccountStatus
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *auditmanager.Client) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
	}

	resp, err := c.GetAccountStatus(ctx, &auditmanager.GetAccountStatusInput{})
	if err != nil {
		return nil, ErrGetAccountStatusFailed
	}

	details := HealthCheck{
		Region:        meta.Region,
		RoleARN:       meta.RoleARN,
		AccountStatus: string(resp.Status),
	}

	if meta.AccountID != "" {
		details.AccountID = meta.AccountID
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
