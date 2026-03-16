package awsassets

import (
	"context"
	"encoding/json"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an AWS Assets health check
type HealthCheck struct {
	// Region is the AWS region used for the session
	Region string `json:"region,omitempty"`
	// RoleARN is the assumed role ARN when present
	RoleARN string `json:"roleArn,omitempty"`
	// AccountID is the AWS account identifier
	AccountID string `json:"accountId,omitempty"`
	// ARN is the caller identity ARN
	ARN string `json:"arn,omitempty"`
	// UserID is the caller identity user identifier
	UserID string `json:"userId,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

// Run validates AWS credentials using STS GetCallerIdentity
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *sts.Client) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
	}

	resp, err := c.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, ErrCallerIdentityLookupFailed
	}

	details := HealthCheck{
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

	return providerkit.EncodeResult(details, ErrResultEncode)
}
