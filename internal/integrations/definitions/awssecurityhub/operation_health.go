package awssecurityhub

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of an AWS Security Hub health check
type HealthCheck struct {
	// HubARN is the Security Hub ARN
	HubARN string `json:"hubArn,omitempty"`
	// SubscribedAt is the Security Hub subscription timestamp
	SubscribedAt string `json:"subscribedAt,omitempty"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(securityHubClient, func(ctx context.Context, request types.OperationRequest, client *securityhub.Client) (json.RawMessage, error) {
		return h.Run(ctx, client)
	})
}

// Run validates Security Hub access by calling DescribeHub
func (HealthCheck) Run(ctx context.Context, c *securityhub.Client) (json.RawMessage, error) {
	resp, err := c.DescribeHub(ctx, &securityhub.DescribeHubInput{})
	if err != nil {
		return nil, ErrDescribeHubFailed
	}

	details := HealthCheck{}

	if resp.HubArn != nil {
		details.HubARN = *resp.HubArn
	}

	if resp.SubscribedAt != nil {
		details.SubscribedAt = *resp.SubscribedAt
	}

	return providerkit.EncodeResult(details, ErrResultEncode)
}
