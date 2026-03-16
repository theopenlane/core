package awsassets

import (
	"context"
	"encoding/json"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
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

// AssetCollect collects AWS asset inventory
type AssetCollect struct {
	// AccountID is the AWS account identifier
	AccountID string `json:"accountId"`
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// Message describes the collection readiness state
	Message string `json:"message"`
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
		return nil, err
	}

	resp, err := c.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("awsassets: STS GetCallerIdentity failed: %w", err)
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

	return jsonx.ToRawMessage(details)
}

// Handle adapts asset collection to the generic operation registration boundary
func (a AssetCollect) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return a.Run(ctx, request.Credential, c)
	}
}

// Run collects AWS asset inventory
func (AssetCollect) Run(ctx context.Context, credential types.CredentialSet, c *sts.Client) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, awsDefaultSessionName)
	if err != nil {
		return nil, err
	}

	resp, err := c.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("awsassets: identity verification failed: %w", err)
	}

	return jsonx.ToRawMessage(AssetCollect{
		AccountID: awssdk.ToString(resp.Account),
		Region:    meta.Region,
		Message:   "aws asset collection ready",
	})
}
