package testutils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/theopenlane/core/internal/graphapi"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/pkg/cp"
	"github.com/theopenlane/core/pkg/middleware/auth"
	mock_shared "github.com/theopenlane/core/pkg/objects/mocks"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/core/pkg/openlaneclient"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/tokens"
	"github.com/vektah/gqlparser/v2/ast"

	ent "github.com/theopenlane/core/internal/ent/generated"
)

var (
	MaxResultLimit         = 10
	TrustCenterCnameTarget = "cname.test.net"
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
func TestClient(c *ent.Client, opts ...openlaneclient.ClientOption) (*testclient.TestClient, error) {
	service, err := MockStorageService(nil, nil)
	if err != nil {
		return nil, err
	}
	e := testEchoServer(c, service, false)

	// setup interceptors
	if opts == nil {
		opts = []openlaneclient.ClientOption{}
	}

	opts = append(opts, openlaneclient.WithTransport(localRoundTripper{server: e}))

	config := openlaneclient.NewDefaultConfig()

	return testclient.New(config, opts...)
}

// TestRestClient creates a new OpenlaneClient for testing
func TestRestClient(c *ent.Client, opts ...openlaneclient.ClientOption) (*openlaneclient.OpenlaneClient, error) {
	service, err := MockStorageService(nil, nil)
	if err != nil {
		return nil, err
	}
	e := testEchoServer(c, service, false)

	// setup interceptors
	if opts == nil {
		opts = []openlaneclient.ClientOption{}
	}

	opts = append(opts, openlaneclient.WithTransport(localRoundTripper{server: e}))

	config := openlaneclient.NewDefaultConfig()

	return openlaneclient.New(config, opts...)
}

// TestClientWithAuth creates a new OpenlaneClient for testing that includes the auth middleware
func TestClientWithAuth(c *ent.Client, opts ...openlaneclient.ClientOption) (*testclient.TestClient, error) {
	service, err := MockStorageService(nil, nil)
	if err != nil {
		return nil, err
	}
	e := testEchoServer(c, service, true)

	// setup interceptors
	if opts == nil {
		opts = []openlaneclient.ClientOption{}
	}

	opts = append(opts, openlaneclient.WithTransport(localRoundTripper{server: e}))

	config := openlaneclient.NewDefaultConfig()

	return testclient.New(config, opts...)
}

// testEchoServer creates a new echo server for testing the graph api
// and optionally includes the middleware for authentication testing
func testEchoServer(c *ent.Client, u *objects.Service, includeMiddleware bool) *echo.Echo {
	srv := testGraphServer(c, u)

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
func createAuthConfig(c *ent.Client) *auth.Options {
	// setup auth middleware
	opts := []auth.Option{
		auth.WithDBClient(c),
		auth.WithValidator(&tokens.MockValidator{}),
	}

	authConfig := auth.NewAuthOptions(opts...)

	return &authConfig
}

// testGraphServer creates a new graphql server for testing the graph api
func testGraphServer(c *ent.Client, u *objects.Service) *handler.Server {
	r := graphapi.NewResolver(c, u).
		WithMaxResultLimit(MaxResultLimit).
		WithTrustCenterCnameTarget(TrustCenterCnameTarget)

	// add the pool to the resolver without a metrics collector
	r.WithPool(100, false) //nolint:mnd

	srv := handler.New(
		gqlgenerated.NewExecutableSchema(
			gqlgenerated.Config{Resolvers: r},
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

	// add metrics middleware to the server
	graphapi.WithMetrics(srv)

	// add the file uploader middleware to the server
	if u != nil {
		graphapi.WithFileUploader(srv, u)
	}

	// if you do not want sleeps (the writer prefers naps anyways), skip cache
	graphapi.WithSkipCache(srv)

	graphapi.WithResultLimit(srv, &MaxResultLimit)

	// Set the error presenter to use the custom error presenter
	srv.SetErrorPresenter(gqlerrors.ErrorPresenter)

	return srv
}

// MockStorageService creates a new storage service for testing with a mock storage backend
func MockStorageService(t *testing.T, uploader storage.UploaderFunc) (*objects.Service, error) {
	return MockStorageServiceWithValidation(t, uploader, nil)
}

// MockStorageServiceWithValidation creates a new storage service for testing with custom validation
func MockStorageServiceWithValidation(t *testing.T, uploader storage.UploaderFunc, validationFunc storage.ValidationFunc) (*objects.Service, error) {
	storageService, _, err := MockStorageServiceWithValidationAndProvider(t, uploader, validationFunc)
	return storageService, err
}

// MockStorageServiceWithValidationAndProvider creates a new storage service for testing with custom validation
// and returns both the StorageService and the mock provider for setting up expectations
func MockStorageServiceWithValidationAndProvider(t *testing.T, uploader storage.UploaderFunc, validationFunc storage.ValidationFunc) (*objects.Service, *mock_shared.MockProvider, error) {
	// Create mock provider - handle nil testing.T gracefully
	var mockProvider *mock_shared.MockProvider
	if t != nil {
		mockProvider = mock_shared.NewMockProvider(t)
	} else {
		// Create a basic mock without test cleanup for non-test contexts
		mockProvider = &mock_shared.MockProvider{}
	}

	// Create cp components
	pool := cp.NewClientPool[storage.Provider](time.Minute)
	clientService := cp.NewClientService(pool)

	// Register mock provider builder
	mockBuilder := &testProviderBuilder{
		provider: mockProvider,
	}
	clientService.RegisterBuilder(cp.ProviderType("mock"), mockBuilder)

	// Create resolver with default rule that selects mock provider
	resolver := cp.NewResolver[storage.Provider]()

	// Add default rule that always returns mock provider
	defaultRule := cp.DefaultRule[storage.Provider](cp.Resolution{
		ClientType:  cp.ProviderType("mock"),
		Credentials: map[string]string{"type": "mock"},
		Config:      map[string]any{"validation": validationFunc != nil},
	})
	resolver.SetDefaultRule(defaultRule)

	// Create objects.Service - simplified for tests
	service := objects.NewService(resolver, clientService)

	// Return service and provider for test setup
	return service, mockProvider, nil
}

// testProviderBuilder implements ClientBuilder for mock providers
type testProviderBuilder struct {
	provider *mock_shared.MockProvider
}

func (b *testProviderBuilder) WithCredentials(credentials map[string]string) cp.ClientBuilder[storage.Provider] {
	return b
}

func (b *testProviderBuilder) WithConfig(config map[string]any) cp.ClientBuilder[storage.Provider] {
	return b
}

func (b *testProviderBuilder) Build(ctx context.Context) (storage.Provider, error) {
	return b.provider, nil
}

func (b *testProviderBuilder) ClientType() cp.ProviderType {
	return cp.ProviderType("mock")
}
