package azureentraid

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// directoryProbeReason is the user-facing reason recorded when the directory probe fails
const directoryProbeReason = "the connection cannot read the Azure Entra ID directory; grant admin consent for the required Microsoft Graph permissions"

// probeDirectory verifies Graph permissions can read the directory
func probeDirectory(ctx context.Context, _ types.OperationRequest, c *msgraphsdk.GraphServiceClient) error {
	_, err := c.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{Top: new(int32(1))},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory probe failed")

		return types.Degraded(ErrUsersFetchFailed, directoryProbeReason)
	}

	return nil
}
