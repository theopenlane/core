package scim

import (
	"context"
	"errors"
	"net/http"

	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

var requestKey = contextx.NewKey[*Request]()

var (
	// ErrSCIMRequestRequired is returned when SCIM request context is missing
	ErrSCIMRequestRequired = errors.New("scim request context is required")
)

// Request carries the resolved installation context for SCIM request processing
type Request struct {
	// Installation is the resolved SCIM integration installation
	Installation *generated.Integration
	// BasePath is the stable SCIM route prefix for this installation, ending in /v2
	BasePath string
}

// WithRequest stores the SCIM request context
func WithRequest(ctx context.Context, req *Request) context.Context {
	return requestKey.Set(ctx, req)
}

// RequestFromContext retrieves the SCIM request context
func RequestFromContext(ctx context.Context) (*Request, bool) {
	return requestKey.Get(ctx)
}

// ResolveRequest extracts the transaction client and SCIM request from the HTTP request context
func ResolveRequest(r *http.Request) (context.Context, *generated.Client, *Request, error) {
	ctx := r.Context()
	client := transaction.FromContext(ctx).Client()

	sr, ok := RequestFromContext(ctx)
	if !ok {
		return ctx, nil, nil, ErrSCIMRequestRequired
	}

	return ctx, client, sr, nil
}
