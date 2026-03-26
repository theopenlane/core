package scim

import (
	"context"
	"errors"
	"net/http"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

var scimRequestKey = contextx.NewKey[*SCIMRequest]()

var (
	// ErrSCIMRequestRequired is returned when SCIM request context is missing
	ErrSCIMRequestRequired = errors.New("scim request context is required")
)

// SCIMRequest carries the resolved installation and runtime for SCIM request processing
type SCIMRequest struct {
	// Installation is the resolved SCIM integration installation
	Installation *generated.Integration
	// Runtime provides shared integration execution capabilities
	Runtime *integrationsruntime.Runtime
	// BasePath is the stable SCIM route prefix for this installation, ending in /v2
	BasePath string
}

// WithSCIMRequest stores the SCIM request context
func WithSCIMRequest(ctx context.Context, req *SCIMRequest) context.Context {
	return scimRequestKey.Set(ctx, req)
}

// SCIMRequestFromContext retrieves the SCIM request context
func SCIMRequestFromContext(ctx context.Context) (*SCIMRequest, bool) {
	return scimRequestKey.Get(ctx)
}

// ResolveRequest extracts the transaction client and SCIM request from the HTTP request context
func ResolveRequest(r *http.Request) (context.Context, *generated.Client, *SCIMRequest, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	sr, ok := SCIMRequestFromContext(ctx)
	if !ok {
		return ctx, nil, nil, ErrSCIMRequestRequired
	}

	return ctx, client, sr, nil
}
