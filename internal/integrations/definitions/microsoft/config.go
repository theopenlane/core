package microsoft

import (
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// AppConfig holds operator-level credentials for one multi-tenant Azure application registration
type AppConfig struct {
	// ApplicationID is the Application (client) ID from the Azure app registration Overview page, used as the OAuth client ID and in the well-known configuration for domain validation
	ApplicationID string `json:"applicationid" koanf:"applicationid"`
	// ClientSecret is the client secret Value (not the Secret ID) from the Azure app registration
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
	// RedirectURL is the OAuth callback URL registered with the Azure application
	RedirectURL string `json:"redirecturl" koanf:"redirecturl" default:"https://api.theopenlane.io/v1/integrations/auth/callback"`
	// DefaultTenant pins the admin consent flow to a specific tenant ID or domain (e.g. for local dev/testing)
	// when empty the generic /organizations endpoint is used
	DefaultTenant string `json:"defaulttenant" koanf:"defaulttenant"`

	// endpoints redirects Microsoft endpoints, only set by the mock provider in tests
	endpoints *endpointOverrides
}

// LegacyApp holds the retired Azure app registration that installations created before the app moved tenants still use
type LegacyApp struct {
	// ApplicationID is the Application (client) ID of the retired app registration
	ApplicationID string `json:"applicationid" koanf:"applicationid"`
	// ClientSecret is the client secret Value of the retired app registration
	ClientSecret string `json:"clientsecret" koanf:"clientsecret" sensitive:"true"`
}

// ForApplication returns the app config for the registration an installation was created with, falling back to legacy when none was recorded
func (c AppConfig) ForApplication(applicationID string, legacy LegacyApp) AppConfig {
	if legacy.ApplicationID == "" {
		return c
	}

	if applicationID == "" || strings.EqualFold(applicationID, legacy.ApplicationID) {
		c.ApplicationID = legacy.ApplicationID
		c.ClientSecret = legacy.ClientSecret
	}

	return c
}

// endpointOverrides points Microsoft clients at alternate login, ARM, and Graph hosts
type endpointOverrides struct {
	// cloud holds the login authority host and ARM endpoint
	cloud cloud.Configuration
	// graphURL is the Graph base URL including the version segment
	graphURL string
	// httpClient sends every request to the overridden endpoints
	httpClient *http.Client
}

// clientOptions returns the azcore client options for the configured endpoints
func (c AppConfig) clientOptions() policy.ClientOptions {
	if c.endpoints == nil {
		return policy.ClientOptions{}
	}

	return policy.ClientOptions{
		Cloud:     c.endpoints.cloud,
		Transport: c.endpoints.httpClient,
		Retry:     policy.RetryOptions{MaxRetries: -1},
	}
}

// ARMClientOptions returns the ARM client options for the configured endpoints, nil for the public cloud
func (c AppConfig) ARMClientOptions() *arm.ClientOptions {
	if c.endpoints == nil {
		return nil
	}

	return &arm.ClientOptions{ClientOptions: c.clientOptions()}
}
