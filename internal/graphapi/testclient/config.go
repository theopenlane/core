package testclient

import (
	"net/url"

	"github.com/Yamashou/gqlgenc/clientv2"
)

// Config is the configuration for the API client
type Config struct {
	// BaseURL is the base URL for the API
	BaseURL *url.URL `json:"baseUrl" yaml:"base_url" default:"https://api.theopenlane.io"`
	// GraphQLPath is the path to the GraphQL endpoint
	GraphQLPath string `json:"graphqlPath" default:"/query"`
	// Interceptors are the request interceptors for the graph client
	Interceptors []clientv2.RequestInterceptor
	// Credentials are the credentials for the client
	Credentials Credentials
	// Clientv2Options are the options for the graph client
	Clientv2Options clientv2.Options
}

// GraphRequestPath returns the full URL path for the GraphQL endpoint
func GraphRequestPath(config *Config) string {
	baseurl := config.BaseURL.String()

	return baseurl + config.GraphQLPath
}

const (
	host        = "localhost:17608"
	defaultPath = "/query"
	historyPath = "/history/query"
)

// NewDefaultConfig returns a new default configuration for the API client
// and connecting to the production environment
func NewDefaultConfig() Config {
	c := defaultConfig

	c.BaseURL = &url.URL{
		Scheme: "https",
		Host:   host,
	}

	return c
}

var defaultConfig = Config{
	BaseURL:         &url.URL{},
	GraphQLPath:     defaultPath,
	Clientv2Options: clientv2.Options{ParseDataAlongWithErrors: true},
}
