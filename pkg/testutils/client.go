package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/theopenlane/core/internal/graphapi"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/core/pkg/objects"
	mock_objects "github.com/theopenlane/core/pkg/objects/mocks"
	"github.com/theopenlane/core/pkg/openlaneclient"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/vektah/gqlparser/v2/ast"

	ent "github.com/theopenlane/core/internal/ent/generated"
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

// TestClient creates a new OpenlaneClient for testing
func TestClient(t *testing.T, c *ent.Client, u *objects.Objects, opts ...openlaneclient.ClientOption) (*openlaneclient.OpenlaneClient, error) {
	e := testEchoServer(t, c, u, false)

	// setup interceptors
	if opts == nil {
		opts = []openlaneclient.ClientOption{}
	}

	opts = append(opts, openlaneclient.WithTransport(localRoundTripper{server: e}))

	config := openlaneclient.NewDefaultConfig()

	return openlaneclient.New(config, opts...)
}

// TestClientWithAuth creates a new OpenlaneClient for testing that includes the auth middleware
func TestClientWithAuth(t *testing.T, c *ent.Client, u *objects.Objects, opts ...openlaneclient.ClientOption) (*openlaneclient.OpenlaneClient, error) {
	e := testEchoServer(t, c, u, true)

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
func testEchoServer(t *testing.T, c *ent.Client, u *objects.Objects, includeMiddleware bool) *echo.Echo {
	srv := testGraphServer(t, c, u)

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
func createAuthConfig(c *ent.Client) *auth.AuthOptions {
	// setup auth middleware
	opts := []auth.AuthOption{
		auth.WithDBClient(c),
	}

	authConfig := auth.NewAuthOptions(opts...)

	authConfig.WithLocalValidator()

	return &authConfig
}

// testGraphServer creates a new graphql server for testing the graph api
func testGraphServer(t *testing.T, c *ent.Client, u *objects.Objects) *handler.Server {
	srv := handler.New(
		gqlgenerated.NewExecutableSchema(
			gqlgenerated.Config{Resolvers: graphapi.NewResolver(c, u)},
		))

	// add all the transports to the server
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// lower the cache size for testing
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100), //nolint:mnd
	})

	// add all extension to the server
	graphapi.AddAllExtensions(srv)

	graphapi.WithTransactions(srv, c)

	// add the file uploader middleware to the server
	if u != nil {
		graphapi.WithFileUploader(srv, u)
	}

	// if you do not want sleeps (the writer prefers naps anyways), skip cache
	graphapi.WithSkipCache(srv)

	return srv
}

// MockObjectManager creates a new objects manager for testing with a mock storage backend
func MockObjectManager(t *testing.T, uploader objects.UploaderFunc) (*objects.Objects, error) {
	return objects.New(
		objects.WithStorage(mock_objects.NewMockStorage(t)),
		objects.WithUploaderFunc(uploader),
	)
}
