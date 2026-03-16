package types

import openapi "github.com/theopenlane/core/common/openapi"

// OAuthPublicConfig holds the public OAuth configuration for a definition
type OAuthPublicConfig = openapi.OAuthPublicConfig

// IntegrationConfig is the per-installation runtime configuration
type IntegrationConfig = openapi.IntegrationConfig

// IntegrationProviderState stores provider-specific state captured during auth and config
type IntegrationProviderState = openapi.IntegrationProviderState
