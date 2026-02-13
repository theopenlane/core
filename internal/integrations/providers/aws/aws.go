package aws

import (
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awssts"
)

// TypeAWS identifies the consolidated AWS provider.
const TypeAWS = types.ProviderType("aws")

// Builder returns the AWS provider builder.
func Builder() providers.Builder {
	return awssts.Builder(
		TypeAWS,
		awssts.WithOperations(awsOperations()),
		awssts.WithClientDescriptors(awsClientDescriptors()),
	)
}
