package corejobs

import (
	"net/url"
	"strings"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

type OpenlaneConfig struct {
	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`
}

var allowedPrefixes = []string{
	// personal access tokens
	"tolp_",
	// api tokens
	"tola_",
	// job runner tokens
	"runner_",
}

func (c *OpenlaneConfig) Validate() error {
	if c.OpenlaneAPIHost == "" {
		return ErrOpenlaneHostMissing
	}

	if c.OpenlaneAPIToken == "" {
		return ErrOpenlaneTokenMissing
	}

	if !validateTokenPrefix(c.OpenlaneAPIToken) {
		return ErrOpenlaneTokenMissing
	}

	return nil
}

func validateTokenPrefix(token string) bool {
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}
	return false
}

// getOpenlaneClient creates and returns a new Openlane client using the provided configuration.
// It configures the client with the appropriate base URL and authentication credentials.
func (c *OpenlaneConfig) getOpenlaneClient() (olclient.OpenlaneClient, error) {
	// validate config
	if err := c.Validate(); err != nil {
		return nil, err
	}

	olconfig := openlaneclient.NewDefaultConfig()

	baseURL, err := url.Parse(c.OpenlaneAPIHost)
	if err != nil {
		return nil, err
	}

	opts := []openlaneclient.ClientOption{openlaneclient.WithBaseURL(baseURL)}

	opts = append(opts, openlaneclient.WithCredentials(openlaneclient.Authorization{
		BearerToken: c.OpenlaneAPIToken,
	}))

	return openlaneclient.New(olconfig, opts...)
}
