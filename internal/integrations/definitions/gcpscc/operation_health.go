package gcpscc

import (
	"context"
	"encoding/json"
	"errors"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/api/iterator"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HealthCheck holds the result of a GCP SCC health check
type HealthCheck struct {
	// Parents is the list of SCC parent resources that were probed
	Parents []string `json:"parents"`
}

// checkHealth executes the GCP SCC health check
func checkHealth(ctx context.Context, request types.OperationRequest, c *cloudscc.Client) (json.RawMessage, error) {
	scope, err := resolveScope(request.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error attempting to resolve credentials")
		return nil, err
	}

	parents, err := resolveParents(scope)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error attempting to resolve parents")
		return nil, err
	}

	for _, parent := range parents {
		req := &securitycenterpb.ListSourcesRequest{
			Parent:   parent,
			PageSize: 1,
		}

		it := c.ListSources(ctx, req)
		_, err = it.Next()

		if errors.Is(err, iterator.Done) {
			err = nil
		}

		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: failed to list sources")
			return nil, ErrListSourcesFailed
		}
	}

	return providerkit.EncodeResult(HealthCheck{Parents: parents}, ErrResultEncode)
}
