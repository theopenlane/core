package handlers

import (
	"github.com/theopenlane/utils/rout"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ConfigureIntegrationRequest is the request type for configuring a non-OAuth provider.
type ConfigureIntegrationRequest = openapi.ConfigureIntegrationRequest

// RunIntegrationOperationBody is the request body for triggering a provider operation.
type RunIntegrationOperationBody = openapi.RunIntegrationOperationBody

// RunIntegrationOperationRequest is the request type for running an integration operation.
type RunIntegrationOperationRequest = openapi.RunIntegrationOperationRequest

// ConfigureIntegrationResponse is the response after successfully configuring a provider.
type ConfigureIntegrationResponse = openapi.ConfigureIntegrationResponse

// RunIntegrationOperationResponse is the response after executing or queuing a provider operation.
type RunIntegrationOperationResponse = openapi.RunIntegrationOperationResponse

// IntegrationProvidersResponse is the response listing available integration definitions.
type IntegrationProvidersResponse struct {
	rout.Reply
	// Providers is the list of available integration definitions.
	Providers []types.Definition `json:"providers"`
}

// IntegrationAuthStartRequest is the request type for starting an integration auth flow.
type IntegrationAuthStartRequest = openapi.IntegrationAuthStartRequest

// ExampleIntegrationAuthStartRequest is an example auth start request for OpenAPI documentation.
var ExampleIntegrationAuthStartRequest = openapi.ExampleIntegrationAuthStartRequest

// ExampleConfigureIntegrationRequest is an example configuration payload for OpenAPI documentation.
var ExampleConfigureIntegrationRequest = openapi.ExampleConfigureIntegrationRequest

// ExampleRunIntegrationOperationRequest is an example operation payload for OpenAPI documentation.
var ExampleRunIntegrationOperationRequest = openapi.ExampleRunIntegrationOperationRequest
