package awssecurityhub

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awssts"
)

// TypeAWSSecurityHub identifies the AWS Security Hub provider.
const TypeAWSSecurityHub = types.ProviderType("aws_security_hub")

// Builder returns the AWS Security Hub provider builder.
func Builder() providers.Builder {
	return awssts.Builder(
		TypeAWSSecurityHub,
		awssts.WithOperations(awsSecurityHubOperations()),
		awssts.WithClientDescriptors(awsSecurityHubClientDescriptors()),
	)
}
