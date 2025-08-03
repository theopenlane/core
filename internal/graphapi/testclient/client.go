package testclient

import (
	"github.com/theopenlane/core/pkg/openlaneclient"
)

type TestClient struct {
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
