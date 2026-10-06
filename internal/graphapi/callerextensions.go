package graphapi

import (
	"github.com/99designs/gqlgen/graphql/handler"
)

// WithCallerExtensions registers the extensions that restrict what specific callers can do, shared by the server and test setup
func WithCallerExtensions(srv *handler.Server) {
	// prevent synthetic support-user list edges from falling through to QueryXxx()
	WithSupportUserEdges(srv)

	// anonymous callers can never use bulk or delete mutations, and may only set entx.AnonymousFields fields
	WithAnonymousMutationBlock(srv)
}
