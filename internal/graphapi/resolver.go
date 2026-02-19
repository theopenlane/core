package graphapi

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	echo "github.com/theopenlane/echox"
	"github.com/vektah/gqlparser/v2/ast"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/directives"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/graphsubscriptions"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/workflows"
	mwauth "github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/core/pkg/gala"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const (
	defaultWebsocketPingInterval = 30 * time.Second
	defaultSSEKeepAliveInterval  = 15 * time.Second
	defaultInitTimeout           = 10 * time.Second
)

// Resolver provides a graph response resolver
type Resolver struct {
	db                *ent.Client
	pool              *gala.Pool
	extensionsEnabled bool
	uploader          *objects.Service
	isDevelopment     bool
	complexityLimit   int
	maxResultLimit    *int
	workflowsConfig   workflows.Config

	// subscription settings
	subscriptionSettings

	// trust center settings
	trustCenterSettings
}

// trustCenterSettings holds the settings for trust center domains
type trustCenterSettings struct {
	// trustCenterCnameTarget is the cname target for trust center domains
	trustCenterCnameTarget string
	// defaultTrustCenterDomain is the default domain for trust center
	defaultTrustCenterDomain string
}

// subscriptionSettings holds the settings for subscriptions
type subscriptionSettings struct {
	// subscriptionsEnabled indicates if subscriptions are turned on
	subscriptionsEnabled bool
	// subscription manager for real-time updates
	subscriptionManager *graphsubscriptions.Manager
	// allowed origins for websocket connections
	origins map[string]struct{}
	// websocketPingInterval is the interval for sending pings to keep the websocket alive
	websocketPingInterval time.Duration
	// sseKeepAliveInterval is the interval for sending keep-alive messages for sse connections
	sseKeepAliveInterval time.Duration
	// authOptions for authenticating websocket connections
	authOptions *mwauth.Options
}

// NewResolver returns a resolver configured with the given ent client
func NewResolver(db *ent.Client, u *objects.Service) *Resolver {
	defaultWorkflows := workflows.NewDefaultConfig()
	return &Resolver{
		db:              db,
		uploader:        u,
		workflowsConfig: *defaultWorkflows,
		subscriptionSettings: subscriptionSettings{
			websocketPingInterval: defaultWebsocketPingInterval,
			sseKeepAliveInterval:  defaultSSEKeepAliveInterval,
		},
	}
}

// WithSubscriptions enables graphql subscriptions to the server using websockets or sse
func (r Resolver) WithSubscriptions(enabled bool) *Resolver {
	if enabled {
		r.subscriptionManager = graphsubscriptions.NewManager()
		r.subscriptionsEnabled = true
	}

	return &r
}

// Handler is an http handler wrapping a Resolver
type Handler struct {
	r              *Resolver
	graphqlHandler *handler.Server
	middleware     []echo.MiddlewareFunc
}

// Handler returns an http handler for a graph resolver
func (r *Resolver) Handler() *Handler {
	c := &gqlgenerated.Config{Resolvers: r}

	directives.ImplementAllDirectives(c)

	srv := handler.New(gqlgenerated.NewExecutableSchema(
		*c,
	))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: r.uploader.MaxSize(),
		MaxMemory:     common.DefaultMaxMemoryMB << 20, //nolint:mnd,
	})

	srv.AddTransport(r.createSSEClient())
	srv.AddTransport(r.CreateWebsocketClient())

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000)) //nolint:mnd

	// only enable introspection in development mode
	if r.isDevelopment {
		srv.Use(extension.Introspection{})
	}

	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100), //nolint:mnd
	})

	// add complexity limit
	r.WithComplexityLimit(srv)

	// add transactional db client
	WithTransactions(srv, r.db)

	// add context level caching
	common.WithContextLevelCache(srv)

	// add extensions if enabled
	if r.extensionsEnabled {
		common.AddAllExtensions(srv)
	}

	// Set the error presenter to use the custom error presenter
	srv.SetErrorPresenter(gqlerrors.ErrorPresenter)

	// add file uploader if it is configured
	if r.uploader != nil {
		common.WithFileUploader(srv, r.uploader)
	}

	// add max result limits to fields in requests
	common.WithResultLimit(srv, r.maxResultLimit)

	common.WithMetrics(srv)

	h := &Handler{
		r:              r,
		graphqlHandler: srv,
	}

	return h
}

// Handler returns the http.HandlerFunc for the GraphAPI
func (h *Handler) Handler() http.HandlerFunc {
	return h.graphqlHandler.ServeHTTP
}

// Routes for the the server
func (h *Handler) Routes(e *echo.Group) {
	e.Use(h.middleware...)

	// Create the default POST graph endpoint
	e.POST("/"+common.GraphPath, func(c echo.Context) error {
		h.graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Create a GET query endpoint in order to create short queries with a query string
	e.GET("/"+common.GraphPath, func(c echo.Context) error {
		h.graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
