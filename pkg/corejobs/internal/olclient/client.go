package olclient

import (
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// OpenlaneClient is a type alias for openlaneclient.OpenlaneGraphClient
// which provides access to the Openlane API
type OpenlaneClient = openlaneclient.OpenlaneGraphClient

// New creates a new Openlane client with the provided configuration and options.
// It returns an implementation of the OpenlaneClient interface and any error encountered.
func New(config openlaneclient.Config, opts ...openlaneclient.ClientOption) (OpenlaneClient, error) {
	return openlaneclient.New(config, opts...)
}
