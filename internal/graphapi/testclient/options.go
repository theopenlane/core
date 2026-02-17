package testclient

import (
	"net/http"
	"net/url"

	"github.com/gqlgo/gqlgenc/clientv2"

	"github.com/theopenlane/httpsling"
)

// ClientOption allows us to configure the APIv1 client when it is created
type ClientOption func(c *APIv1) error

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
		return c.Requester.Apply(httpsling.BearerAuth(auth.BearerToken))
	}
}

// WithAPIToken sets the API token for the APIv1 client
// This is the recommended way to authenticate using an API token or personal access token
// use WithCredentials for more advanced authentication methods
func WithAPIToken(apiToken string) ClientOption {
	return func(c *APIv1) error {
		creds := Authorization{
			BearerToken: apiToken,
		}

		return WithCredentials(creds)(c)
	}
}

// WithInterceptors sets the interceptors for the APIv1 client
func WithInterceptors(interceptors clientv2.RequestInterceptor) ClientOption {
	return func(c *APIv1) error {
		if c.Config.Interceptors == nil {
			c.Config.Interceptors = []clientv2.RequestInterceptor{}
		}

		c.Config.Interceptors = append(c.Config.Interceptors, interceptors)

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
func WithBaseURL(baseURL string) ClientOption {
	return func(c *APIv1) error {
		url, err := url.Parse(baseURL)
		if err != nil {
			return err
		}

		// Set the base URL for the APIv1 client
		c.Config.BaseURL = url

		// Set the base URL for the HTTPSling client
		return c.Requester.Apply(httpsling.URL(url.String()))
	}
}

// WithTransport sets the transport for the APIv1 client
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *APIv1) error {
		c.Requester.HTTPClient().Transport = transport

		return nil
	}
}

// WithCSRFToken sets the CSRF token header on the client for all requests
func WithCSRFToken(token string) ClientOption {
	return func(c *APIv1) error {
		if token == "" {
			// If the token is empty, we do not set the interceptor
			return nil
		}

		c.Config.Interceptors = append(c.Config.Interceptors, WithCSRFTokenInterceptor(token))

		// set the CSRF token header on the HTTPSling client for REST requests
		return c.Requester.Apply(httpsling.Header(csrfHeader, token))
	}
}
