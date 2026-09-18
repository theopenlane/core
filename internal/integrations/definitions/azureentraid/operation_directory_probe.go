package azureentraid

import (
	"context"
	"encoding/json"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// directoryProbeReason is the user-facing reason recorded when the directory probe fails
const directoryProbeReason = "the connection cannot read the Azure Entra ID directory; grant admin consent for the required Microsoft Graph permissions"

// DirectoryProbe holds the result of the directory sync health probe
type DirectoryProbe struct {
	// DirectoryReadable reports whether the granted Graph permissions can read the directory
	DirectoryReadable bool `json:"directoryReadable"`
}

// Handle adapts the probe to the operation health check boundary
func (p DirectoryProbe) Handle() types.OperationHandler {
	return providerkit.WithClient(entraClient, p.Run)
}

// Run verifies the granted Graph permissions can read the directory by fetching a single user
func (DirectoryProbe) Run(ctx context.Context, c *msgraphsdk.GraphServiceClient) (json.RawMessage, error) {
	_, err := c.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{Top: new(int32(1))},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("directory probe failed")

		return nil, types.Degraded(ErrUsersFetchFailed, directoryProbeReason)
	}

	return providerkit.EncodeResult(DirectoryProbe{DirectoryReadable: true}, ErrResultEncode)
}
