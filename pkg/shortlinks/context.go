package shortlinks

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// suppressed marks a context whose sends must deliver original URLs instead of creating
// links, keeping test and preview sends out of the service
var suppressed = contextx.NewKey[struct{}]()

// ContextWithoutShortlinks returns a context in which Shorten delivers the original URL
func ContextWithoutShortlinks(ctx context.Context) context.Context {
	return suppressed.Set(ctx, struct{}{})
}

// DisabledFromContext reports whether shortlink creation is suppressed for this context
func DisabledFromContext(ctx context.Context) bool {
	_, ok := suppressed.Get(ctx)

	return ok
}
