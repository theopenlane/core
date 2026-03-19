package awssecurityhub

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/definitions/awskit"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an AWS Security Hub health check
type HealthCheck struct {
	// Region is the AWS region used for the session
	Region string `json:"region"`
	// RoleARN is the assumed role ARN when present
	RoleARN string `json:"roleArn,omitempty"`
	// HubARN is the Security Hub ARN
	HubARN string `json:"hubArn,omitempty"`
	// SubscribedAt is the Security Hub subscription timestamp
	SubscribedAt string `json:"subscribedAt,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		c, err := SecurityHubClient.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return h.Run(ctx, request.Credential, c)
	}
}

// Run validates Security Hub access by calling DescribeHub
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *securityhub.Client) (json.RawMessage, error) {
	meta, err := awskit.MetadataFromProviderData(credential.ProviderData, defaultSessionName)
	if err != nil {
		return nil, ErrCredentialMetadataInvalid
	}

	resp, err := c.DescribeHub(ctx, &securityhub.DescribeHubInput{})
	if err != nil {
		return nil, ErrDescribeHubFailed
	}

	details := HealthCheck{
		Region:  meta.Region,
		RoleARN: meta.RoleARN,
	}

	if resp.HubArn != nil {
		details.HubARN = *resp.HubArn
	}

	if resp.SubscribedAt != nil {
		details.SubscribedAt = *resp.SubscribedAt
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
