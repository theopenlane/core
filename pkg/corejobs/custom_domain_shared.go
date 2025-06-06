package corejobs

import (
	"net/url"

	"github.com/theopenlane/core/pkg/openlaneclient"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
)

// CustomDomainConfig contains the configuration for the custom domain workers
type CustomDomainConfig struct {
	CloudflareAPIKey string `koanf:"cloudflareApiKey" json:"cloudflareApiKey" jsonschema:"required description=the cloudflare api key"`

	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`

	DatabaseHost string `koanf:"databaseHost" json:"databaseHost" jsonschema:"required description=the database host"`
}

// getOpenlaneClient creates and returns a new Openlane client using the provided configuration.
// It configures the client with the appropriate base URL and authentication credentials.
func getOpenlaneClient(config CustomDomainConfig) (olclient.OpenlaneClient, error) {
	olconfig := openlaneclient.NewDefaultConfig()

	baseURL, err := url.Parse(config.OpenlaneAPIHost)
	if err != nil {
		return nil, err
	}

	opts := []openlaneclient.ClientOption{openlaneclient.WithBaseURL(baseURL)}
	opts = append(opts, openlaneclient.WithCredentials(openlaneclient.Authorization{
		BearerToken: config.OpenlaneAPIToken,
	}))

	return openlaneclient.New(olconfig, opts...)
}
