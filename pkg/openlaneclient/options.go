package openlaneclient

import (
	"net/http"
	"net/url"

	"github.com/Yamashou/gqlgenc/clientv2"

	"github.com/theopenlane/core/pkg/httpsling"
)

// ClientOption allows us to configure the APIv1 client when it is created
type ClientOption func(c *APIv1) error

// WithClient sets the client for the APIv1 client
func WithClient(client *httpsling.Client) ClientOption {
	return func(c *APIv1) error {
		c.HTTPSlingClient = client
		return nil
	}
}

// WithCredentials sets the credentials for the APIv1 client
func WithCredentials(creds Credentials) ClientOption {
	return func(c *APIv1) error {
		c.Config.Credentials = creds

		// If the credentials are set, we should also set the token for graph interceptor
		auth, err := NewAuthorization(creds)
		if err != nil {
			return err
		}

		c.Config.Interceptors = append(c.Config.Interceptors, auth.WithAuthorization())

		// Set the bearer token for the HTTPSling client, used for REST requests
		c.Config.HTTPSling.Headers.Set(httpsling.HeaderAuthorization, "Bearer "+auth.BearerToken)

		return nil
	}
}

// WithHTTPSlingConfig sets the config for the APIv1 client
func WithHTTPSlingConfig(config *httpsling.Config) ClientOption {
	return func(c *APIv1) error {
		c.Config.HTTPSling = config
		return nil
	}
}

// WithInterceptors sets the interceptors for the APIv1 client
func WithInterceptors(interceptors clientv2.RequestInterceptor) ClientOption {
	return func(c *APIv1) error {
		c.Config.Interceptors = []clientv2.RequestInterceptor{interceptors}
		return nil
	}
}

// WithClientV2Option sets the clientv2 options for the APIv1 client
func WithClientV2Option(option clientv2.Options) ClientOption {
	return func(c *APIv1) error {
		c.Config.Clientv2Options = option
		return nil
	}
}

// WithGraphQueryEndpoint sets the graph query endpoint for the APIv1 client
func WithGraphQueryEndpoint(endpoint string) ClientOption {
	return func(c *APIv1) error {
		c.Config.GraphQLPath = endpoint
		return nil
	}
}

// WithBaseURL sets the base URL for the APIv1 client
func WithBaseURL(baseURL *url.URL) ClientOption {
	return func(c *APIv1) error {
		// Set the base URL for the APIv1 client
		c.Config.BaseURL = baseURL

		// Set the base URL for the HTTPSling client
		c.Config.HTTPSling.BaseURL = baseURL.String()

		return nil
	}
}

// WithTransport sets the transport for the APIv1 client
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *APIv1) error {
		c.Config.HTTPSling.Transport = transport
		return nil
	}
}
