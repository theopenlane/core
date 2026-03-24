package awssecurityhub

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"

	"github.com/theopenlane/core/internal/integrations/types"
)

// SecurityHubClientBuilder builds AWS Security Hub clients for one installation
type SecurityHubClientBuilder struct{}

// Build constructs the AWS Security Hub client using the shared AWS credential inputs
func (SecurityHubClientBuilder) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	return buildAWSServiceClient(ctx, req, func(cfg awssdk.Config) *securityhub.Client {
		return securityhub.NewFromConfig(cfg)
	})
}

// AuditManagerClientBuilder builds AWS Audit Manager clients for one installation
type AuditManagerClientBuilder struct{}

// Build constructs the AWS Audit Manager client using the shared AWS credential inputs
func (AuditManagerClientBuilder) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	return buildAWSServiceClient(ctx, req, func(cfg awssdk.Config) *auditmanager.Client {
		return auditmanager.NewFromConfig(cfg)
	})
}
