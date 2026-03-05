package aws

import (
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/awssts"
	"github.com/theopenlane/core/internal/integrations/types"
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
