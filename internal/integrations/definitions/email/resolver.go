package email

import "context"

// ClientResolver resolves the appropriate EmailClient for a given owner.
// When the owner has a customer email integration installed, it returns
// the customer-provisioned client; otherwise it falls back to the system client
type ClientResolver func(ctx context.Context, ownerID string) (*EmailClient, error)

// clientResolver is the package-level resolver wired at server startup
var clientResolver ClientResolver

// SetClientResolver installs the client resolver used by ResolveClient.
// Must be called once during server initialization after the integration runtime is ready
func SetClientResolver(resolver ClientResolver) {
	clientResolver = resolver
}

// ResolveClient returns the appropriate EmailClient for the given owner.
// Delegates to the resolver installed by SetClientResolver
func ResolveClient(ctx context.Context, ownerID string) (*EmailClient, error) {
	if clientResolver == nil {
		return nil, ErrClientResolverNotConfigured
	}

	return clientResolver(ctx, ownerID)
}
