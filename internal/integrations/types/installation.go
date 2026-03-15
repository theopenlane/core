package types

import "github.com/theopenlane/core/internal/integrations/schema"

// AuthKind identifies the authentication mechanism used by a definition
type AuthKind = schema.AuthKind

const (
	// AuthKindUnknown represents an unset authentication kind
	AuthKindUnknown = schema.AuthKindUnknown
	// AuthKindOAuth2 represents OAuth2 authentication
	AuthKindOAuth2 = schema.AuthKindOAuth2
	// AuthKindOAuth2ClientCredentials represents OAuth2 client-credentials authentication
	AuthKindOAuth2ClientCredentials = schema.AuthKindOAuth2ClientCredentials
	// AuthKindOIDC represents OpenID Connect authentication
	AuthKindOIDC = schema.AuthKindOIDC
	// AuthKindAPIKey represents API key authentication
	AuthKindAPIKey = schema.AuthKindAPIKey
	// AuthKindGitHubApp represents GitHub App authentication
	AuthKindGitHubApp = schema.AuthKindGitHubApp
	// AuthKindWorkloadIdentity represents workload identity authentication
	AuthKindWorkloadIdentity = schema.AuthKindWorkloadIdentity
	// AuthKindAWSFederation represents AWS STS federation authentication
	AuthKindAWSFederation = schema.AuthKindAWSFederation
	// AuthKindNone represents push-based providers where the external system calls us
	AuthKindNone = schema.AuthKindNone
)

// OAuthPublicConfig holds the public OAuth configuration for a definition
type OAuthPublicConfig = schema.OAuthPublicConfig

// PersistenceConfig controls credential storage behaviour for a definition
type PersistenceConfig = schema.PersistenceConfig

// IntegrationProviderMetadata is a snapshot of definition metadata captured on installation
type IntegrationProviderMetadata = schema.IntegrationProviderMetadata

// IntegrationConfig is the per-installation runtime configuration
type IntegrationConfig = schema.IntegrationConfig

// IntegrationProviderState stores provider-specific state captured during auth and config
type IntegrationProviderState = schema.IntegrationProviderState
