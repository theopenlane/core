package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/core/pkg/middleware/echocontext"
	"github.com/theopenlane/core/pkg/openlaneclient"
	echo "github.com/theopenlane/echox"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// localRoundTripper is an http.RoundTripper that executes HTTP transactions
// by using handler directly, instead of going over an HTTP connection.
type localRoundTripper struct {
	server *echo.Echo
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.server.ServeHTTP(w, req)

	return w.Result(), nil
}

// TestClient creates a new DatumClient for testing
func TestClient(t *testing.T, c *generated.Client, opts ...openlaneclient.ClientOption) (*openlaneclient.DatumClient, error) {
	e := testEchoServer(t, c, false)

	// setup interceptors
	if opts == nil {
		opts = []openlaneclient.ClientOption{}
	}

	opts = append(opts, openlaneclient.WithTransport(localRoundTripper{server: e}))

	config := openlaneclient.NewDefaultConfig()

	return openlaneclient.New(config, opts...)
}

// TestClientWithAuth creates a new DatumClient for testing that includes the auth middleware
func TestClientWithAuth(t *testing.T, c *generated.Client, opts ...openlaneclient.ClientOption) (*openlaneclient.DatumClient, error) {
	e := testEchoServer(t, c, true)

	// setup interceptors
	if opts == nil {
		opts = []openlaneclient.ClientOption{}
	}

	opts = append(opts, openlaneclient.WithTransport(localRoundTripper{server: e}))

	config := openlaneclient.NewDefaultConfig()

	return openlaneclient.New(config, opts...)
}

// testEchoServer creates a new echo server for testing the graph api
// and optionally includes the middleware for authentication testing
func testEchoServer(t *testing.T, c *generated.Client, includeMiddleware bool) *echo.Echo {
	srv := testGraphServer(t, c)

	e := echo.New()

	if includeMiddleware {
		e.Use(echocontext.EchoContextToContextMiddleware())
		e.Use(auth.Authenticate(createAuthConfig(c)))
	}

	e.POST("/query", func(c echo.Context) error {
		req := c.Request()
		res := c.Response()

		srv.ServeHTTP(res, req)

		return nil
	})

	return e
}

// createAuthConfig creates a new auth config for testing with the provided client
// and local validator
func createAuthConfig(c *generated.Client) *auth.AuthOptions {
	// setup auth middleware
	opts := []auth.AuthOption{
		auth.WithDBClient(c),
	}

	authConfig := auth.NewAuthOptions(opts...)

	authConfig.WithLocalValidator()

	return &authConfig
}

// testGraphServer creates a new graphql server for testing the graph api
func testGraphServer(t *testing.T, c *generated.Client) *handler.Server {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.ErrorLevel)).Sugar()

	srv := handler.NewDefaultServer(
		graphapi.NewExecutableSchema(
			graphapi.Config{Resolvers: graphapi.NewResolver(c).WithLogger(logger)},
		))

	// lower the cache size for testing
	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100), //nolint:mnd
	})

	// add all extension to the server
	graphapi.AddAllExtensions(srv)

	graphapi.WithTransactions(srv, c)

	// if you do not want sleeps (the writer prefers naps anyways), skip cache
	graphapi.WithSkipCache(srv)

	return srv
}
