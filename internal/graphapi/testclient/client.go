package testclient

import (
	openlaneclient "github.com/theopenlane/go-client"
)

// TestClient includes the generated GraphQL client and the Openlane REST client
type TestClient struct {
	// TestGraphClient is the generated GraphQL client with possibly more complex queries to make
	// tests easier to write than the OpenlaneGraphClient
	TestGraphClient
	openlaneclient.OpenlaneRestClient
}

// New creates a new API v1 client that implements the Client interface
func New(config openlaneclient.Config, opts ...openlaneclient.ClientOption) (*TestClient, error) {
	// configure rest client
	c, err := openlaneclient.NewRestClient(config, opts...)
	if err != nil {
		return nil, err
	}

	api := c.(*openlaneclient.APIv1)

	// create the graph client
	// use api.Config instead of config because some fields are updated in NewRestClient
	graphClient := NewClient(
		api.Requester.HTTPClient(),
		openlaneclient.GraphRequestPath(api.Config),
		&api.Config.Clientv2Options,
		api.Config.Interceptors...,
	)

	return &TestClient{
		TestGraphClient:    graphClient,
		OpenlaneRestClient: c,
	}, nil
}
