package awssecurityhub

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/types"
)

// SecurityHubClientBuilder builds AWS Security Hub clients for one installation
type SecurityHubClientBuilder struct {
	// cfg is the operator-level config holding Openlane's source AWS credentials
	cfg Config
}

// Build constructs the AWS Security Hub client using the shared AWS credential inputs
func (b SecurityHubClientBuilder) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	return buildAWSServiceClient(ctx, b.cfg, req, func(cfg awssdk.Config) *securityhub.Client {
		return securityhub.NewFromConfig(cfg)
	})
}

// ConfigServiceClientBuilder builds AWS Config clients for one installation
type ConfigServiceClientBuilder struct {
	// cfg is the operator-level config holding Openlane's source AWS credentials
	cfg Config
}

// Build constructs the AWS Config client using the shared AWS credential inputs
func (b ConfigServiceClientBuilder) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	return buildAWSServiceClient(ctx, b.cfg, req, func(cfg awssdk.Config) *configservice.Client {
		return configservice.NewFromConfig(cfg)
	})
}

// IAMClientBuilder builds AWS IAM clients for one installation
type IAMClientBuilder struct {
	// cfg is the operator-level config holding Openlane's source AWS credentials
	cfg Config
}

// Build constructs the AWS IAM client using the shared AWS credential inputs
func (b IAMClientBuilder) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	return buildAWSServiceClient(ctx, b.cfg, req, func(cfg awssdk.Config) *iam.Client {
		return iam.NewFromConfig(cfg)
	})
}
