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
	"github.com/labstack/gommon/log"
	"github.com/theopenlane/core/internal/graphapi"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/pkg/gqlerrors"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/middleware/auth"
	mock_shared "github.com/theopenlane/core/pkg/objects/mocks"
	objects "github.com/theopenlane/core/pkg/objects/objstore"
	"github.com/theopenlane/core/pkg/objects/storage"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/eddy"
	"github.com/theopenlane/eddy/helpers"
	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/iam/tokens"
	"github.com/vektah/gqlparser/v2/ast"

	ent "github.com/theopenlane/ent/generated"
)

var (
	MaxResultLimit           = 10
	TrustCenterCnameTarget   = "cname.test.net"
	TrustCenterDefaultDomain = "trust.test.net"
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
func TestClient(c *ent.Client, objectStore *objects.Service, opts ...openlaneclient.ClientOption) (*testclient.TestClient, error) {
	var service *objects.Service
	var err error

	if objectStore != nil {
		service = objectStore
	} else {
		service, err = MockStorageService(nil, nil)
		if err != nil {
			return nil, err
		}
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
func TestClientWithAuth(c *ent.Client, objectStore *objects.Service, opts ...openlaneclient.ClientOption) (*testclient.TestClient, error) {
	var service *objects.Service
	var err error

	if objectStore != nil {
		service = objectStore
	} else {
		service, err = MockStorageService(nil, nil)
		if err != nil {
			return nil, err
		}
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

	zLvl, _ := logx.MatchEchoLevel(log.ERROR)
	loggers := logx.Configure(logx.LoggerConfig{
		Level:         zLvl,
		Pretty:        true,
		IncludeCaller: true,
		WithEcho:      true,
	})
	e.Logger = loggers.Echo

	if includeMiddleware {
		e.Use(echocontext.EchoContextToContextMiddleware())
		e.Use(auth.Authenticate(createAuthConfig(c)))
		e.Use(logx.LoggingMiddleware(logx.Config{
			Logger:          loggers.Echo,
			RequestIDHeader: "X-Request-ID",
			RequestIDKey:    "request_id",
			HandleError:     true,
		}))

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
		WithTrustCenterCnameTarget(TrustCenterCnameTarget).
		WithTrustCenterDefaultDomain(TrustCenterDefaultDomain)

	// add the pool to the resolver without a metrics collector
	r.WithPool(100, false) //nolint:mnd

	conf := gqlgenerated.Config{Resolvers: r}

	graphapi.ImplementAllDirectives(&conf)

	srv := handler.New(
		gqlgenerated.NewExecutableSchema(
			conf,
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

	// Create eddy components
	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService(pool, eddy.WithConfigClone[
		storage.Provider,
		storage.ProviderCredentials](func(in *storage.ProviderOptions) *storage.ProviderOptions {
		if in == nil {
			return nil
		}
		return in.Clone()
	}))

	// Create mock provider builder
	mockBuilder := &testProviderBuilder{
		provider: mockProvider,
	}

	// Create resolver with default rule that selects mock provider
	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()

	// Add default rule that always returns mock provider with builder
	defaultRule := &helpers.ConditionalRule[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Predicate: func(_ context.Context) bool {
			return true
		},
		Resolver: func(_ context.Context) (*eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions], error) {
			return &eddy.ResolvedProvider[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: mockBuilder,
				Output:  storage.ProviderCredentials{},
				Config:  storage.NewProviderOptions(storage.WithExtra("validation", validationFunc != nil)),
			}, nil
		},
	}
	resolver.AddRule(defaultRule)

	// Create objects.Service - simplified for tests
	service := objects.NewService(objects.Config{
		Resolver:       resolver,
		ClientService:  clientService,
		ValidationFunc: validationFunc,
	})

	// Return service and provider for test setup
	return service, mockProvider, nil
}

// testProviderBuilder implements eddy.Builder for mock providers
type testProviderBuilder struct {
	provider *mock_shared.MockProvider
}

func (b *testProviderBuilder) Build(ctx context.Context, creds storage.ProviderCredentials, config *storage.ProviderOptions) (storage.Provider, error) {
	return b.provider, nil
}

func (b *testProviderBuilder) ProviderType() string {
	return "mock"
}
