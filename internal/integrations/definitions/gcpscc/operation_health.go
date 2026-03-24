package gcpscc

import (
	"context"
	"encoding/json"
	"errors"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/api/iterator"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// HealthCheck holds the result of a GCP SCC health check
type HealthCheck struct {
	// Parents is the list of SCC parent resources that were probed
	Parents []string `json:"parents"`
}

// Handle adapts the health check to the generic operation registration boundary
func (h HealthCheck) Handle() types.OperationHandler {
	return providerkit.WithClientRequest(sccClient, func(ctx context.Context, request types.OperationRequest, client *cloudscc.Client) (json.RawMessage, error) {
		return h.Run(ctx, request.Credential, client)
	})
}

// Run executes the GCP SCC health check
func (HealthCheck) Run(ctx context.Context, credential types.CredentialSet, c *cloudscc.Client) (json.RawMessage, error) {
	meta, err := metadataFromCredential(credential)
	if err != nil {
		return nil, err
	}

	parents, err := resolveParents(meta)
	if err != nil {
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
			return nil, ErrListSourcesFailed
		}
	}

	return providerkit.EncodeResult(HealthCheck{Parents: parents}, ErrResultEncode)
}
