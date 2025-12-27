package testclient

import (
	"net/http"

	"github.com/theopenlane/httpsling"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TestClient includes the generated GraphQL client and the Openlane REST client
type TestClient struct {
	// TestGraphClient is the generated GraphQL client with possibly more complex queries to make
	// tests easier to write than the OpenlaneGraphClient
	TestGraphClient
	RestClient
}

// APIv1 implements the Client interface and provides methods to interact with the API
type APIv1 struct {
	// Config is the configuration for the APIv1 client
	Config *Config
	// Requester is the HTTP client for the APIv1 client
	Requester *httpsling.Requester
}

// New creates a new API v1 client that implements the Client interface
func New(config Config, opts ...ClientOption) (*TestClient, error) {
	// configure rest client
	c, err := NewRestClient(config, opts...)
	if err != nil {
		return nil, err
	}

	api := c.(*APIv1)

	// create the graph client
	// use api.Config instead of config because some fields are updated in NewRestClient
	graphClient := NewClient(
		api.Requester.HTTPClient(),
		GraphRequestPath(api.Config),
		&api.Config.Clientv2Options,
		api.Config.Interceptors...,
	)

	return &TestClient{
		TestGraphClient: graphClient,
		RestClient:      c,
	}, nil
}

// Cookies returns the cookies set from the previous request(s) on the client Jar.
func (c *TestClient) Cookies() ([]*http.Cookie, error) {
	if c.HTTPSlingRequester().CookieJar() == nil {
		return nil, ErrNoCookieJarSet
	}

	cookies := c.HTTPSlingRequester().CookieJar().Cookies(c.Config().BaseURL)

	return cookies, nil
}

// HTTPSlingRequester is the http client for the APIv1 client
func (c *TestClient) HTTPSlingRequester() *httpsling.Requester {
	api := c.RestClient.(*APIv1)

	return api.Requester
}

// Config is the configuration for the APIv1 client
func (c *TestClient) Config() *Config {
	api := c.RestClient.(*APIv1)

	return api.Config
}

// GetErrorCode returns the error code from the GraphQL error extensions
func GetErrorCode(err error) string {
	if err == nil {
		return ""
	}

	gqlErr, ok := err.(*gqlerror.Error)
	if !ok {
		return ""
	}

	if code, ok := gqlErr.Extensions["code"].(string); ok {
		return code
	}

	return ""
}

// GetErrorMessage returns the error message from the GraphQL error extensions
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	gqlErr, ok := err.(*gqlerror.Error)
	if !ok {
		return ""
	}

	if message, ok := gqlErr.Extensions["message"].(string); ok {
		return message
	}

	return ""
}
