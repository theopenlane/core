package openlaneclient

import (
	"net/url"

	"github.com/Yamashou/gqlgenc/clientv2"
)

// Config is the configuration for the API client
type Config struct {
	// BaseURL is the base URL for the API
	BaseURL *url.URL `json:"baseUrl" yaml:"base_url" default:"http://localhost:17608"`
	// GraphQLPath is the path to the GraphQL endpoint
	GraphQLPath string `json:"graphqlPath" default:"/query"`
	// Interceptors are the request interceptors for the graph client
	Interceptors []clientv2.RequestInterceptor
	// Credentials are the credentials for the client
	Credentials Credentials
	// Clientv2Options are the options for the graph client
	Clientv2Options clientv2.Options
}

// graphRequestPath returns the full URL path for the GraphQL endpoint
func graphRequestPath(config Config) string {
	baseurl := config.BaseURL.String()

	return baseurl + config.GraphQLPath
}

// NewDefaultConfig returns a new default configuration for the API client
func NewDefaultConfig() Config {
	return defaultClientConfig
}

var defaultClientConfig = Config{
	BaseURL: &url.URL{
		Scheme: "http",
		Host:   "localhost:17608",
	},
	GraphQLPath:     "/query",
	Interceptors:    []clientv2.RequestInterceptor{},
	Clientv2Options: clientv2.Options{ParseDataAlongWithErrors: false},
}
