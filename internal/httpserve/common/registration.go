//revive:disable:var-naming
package common

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

var registrationMarkerContextKey = contextx.NewKey[struct{}]()

// WithRegistrationMarker marks a context as OpenAPI registration context.
func WithRegistrationMarker(ctx context.Context) context.Context {
	return registrationMarkerContextKey.Set(ctx, struct{}{})
}

// IsRegistrationContext reports whether the context is marked for OpenAPI registration.
func IsRegistrationContext(ctx context.Context) bool {
	_, ok := registrationMarkerContextKey.Get(ctx)
	return ok
}
