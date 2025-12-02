package graphapi

import (
	"context"
	"net/http"
	"time"

	"ariga.io/entcache"
	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/alitto/pond/v2"
	"github.com/gorilla/websocket"
	"github.com/ravilushqa/otelgqlgen"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/vektah/gqlparser/v2/ast"

	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/internal/graphsubscriptions"
	"github.com/theopenlane/core/pkg/directives"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/gqlerrors"
	"github.com/theopenlane/core/pkg/objects/objstore"
	objects "github.com/theopenlane/core/pkg/objects/objstore"
	ent "github.com/theopenlane/ent/generated"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const (
	ActionGet    = "get"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionCreate = "create"

	// DefaultMaxMemoryMB is the default max memory for multipart forms (32MB)
	DefaultMaxMemoryMB = 32
)

var (
	graphPath               = "query"
	defaultComplexityLimit  = 100
	introspectionComplexity = 200
)

// Resolver provides a graph response resolver
type Resolver struct {
	db                *ent.Client
	pool              *soiree.PondPool
	extensionsEnabled bool
	uploader          *objstore.Service
	isDevelopment     bool
	complexityLimit   int
	maxResultLimit    *int
	// mappable domain that trust center records will resolve to
	trustCenterCnameTarget   string
	defaultTrustCenterDomain string
	// subscription manager for real-time updates
	subscriptionManager *graphsubscriptions.Manager
}

// NewResolver returns a resolver configured with the given ent client
func NewResolver(db *ent.Client, u *objstore.Service) *Resolver {
	return &Resolver{
		db:       db,
		uploader: u,
	}
}

// WithSubscriptions enables graphql subscriptions to the server using websockets or sse
func (r Resolver) WithSubscriptions(enabled bool) *Resolver {
	if enabled {
		r.subscriptionManager = graphsubscriptions.NewManager()
	}

	return &r
}

func (r Resolver) WithTrustCenterCnameTarget(cname string) *Resolver {
	r.trustCenterCnameTarget = cname

	return &r
}

func (r Resolver) WithTrustCenterDefaultDomain(domain string) *Resolver {
	r.defaultTrustCenterDomain = domain

	return &r
}

func (r Resolver) WithExtensions(enabled bool) *Resolver {
	r.extensionsEnabled = enabled

	return &r
}

// WithDevelopment sets the resolver to development mode
// when isDevelopment is false, introspection will be disabled
func (r Resolver) WithDevelopment(dev bool) *Resolver {
	r.isDevelopment = dev

	return &r
}

// WithComplexityLimitConfig sets the complexity limit for the resolver
func (r Resolver) WithComplexityLimitConfig(limit int) *Resolver {
	r.complexityLimit = limit

	return &r
}

// WithMaxResultLimit sets the max result limit in the config for the resolvers
func (r Resolver) WithMaxResultLimit(limit int) *Resolver {
	r.maxResultLimit = &limit

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

	ImplementAllDirectives(c)

	srv := handler.New(gqlgenerated.NewExecutableSchema(
		*c,
	))

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second, //nolint:mnd
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true
			},
		},
	})
	if r.subscriptionManager != nil {
		srv.AddTransport(transport.SSE{
			KeepAlivePingInterval: 10 * time.Second, //nolint:mnd
		})
	}
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: r.uploader.MaxSize(),
		MaxMemory:     DefaultMaxMemoryMB << 20, //nolint:mnd,
	})

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
	WithContextLevelCache(srv)

	// add extensions if enabled
	if r.extensionsEnabled {
		AddAllExtensions(srv)
	}

	// Set the error presenter to use the custom error presenter
	srv.SetErrorPresenter(gqlerrors.ErrorPresenter)

	// add file uploader if it is configured
	if r.uploader != nil {
		WithFileUploader(srv, r.uploader)
	}

	// add max result limits to fields in requests
	WithResultLimit(srv, r.maxResultLimit)

	WithMetrics(srv)

	srv.Use(otelgqlgen.Middleware())

	h := &Handler{
		r:              r,
		graphqlHandler: srv,
	}

	return h
}

// ImplementAllDirectives is a helper function that can be used to add all active directives to the gqlgen config
// in the resolver setup
func ImplementAllDirectives(cfg *gqlgenerated.Config) {
	cfg.Directives.Hidden = directives.HiddenDirective
	cfg.Directives.ReadOnly = directives.ReadOnlyDirective
	cfg.Directives.ExternalReadOnly = directives.ExternalReadOnlyDirective
	cfg.Directives.ExternalSource = directives.ExternalSourceDirective
}

func (r *Resolver) WithComplexityLimit(h *handler.Server) {
	// prevent complex queries except the introspection query
	h.Use(newComplexityLimitWithMetrics(func(_ context.Context, rc *graphql.OperationContext) int {
		if rc != nil && rc.OperationName == "IntrospectionQuery" {
			return introspectionComplexity
		}

		if rc.OperationName == "GlobalSearch" {
			// allow more complexity for the global search
			// e.g. if the complexity limit is 100, we allow 500 for the global search
			return r.complexityLimit * 5 //nolint:mnd
		}

		if r.complexityLimit > 0 {
			return r.complexityLimit
		}

		return defaultComplexityLimit
	}))
}

// WithTransactions adds the transactioner to the ent db client
func WithTransactions(h *handler.Server, d *ent.Client) {
	// setup transactional db client
	h.AroundOperations(injectClient(d))

	h.Use(entgql.Transactioner{TxOpener: d})
}

// WithFileUploader adds the file uploader to the graphql handler
// this will handle the file upload process for the multipart form
func WithFileUploader(h *handler.Server, u *objects.Service) {
	h.AroundFields(injectFileUploader(u))
}

// WithContextLevelCache adds a context level cache to the handler
func WithContextLevelCache(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		if op := graphql.GetOperationContext(ctx).Operation; op != nil && op.Operation == ast.Query {
			ctx = entcache.NewContext(ctx)
		}

		return next(ctx)
	})
}

// WithResultLimit adds a max result limit to the handler in order to set limits on
// all nested edges in the graphql request
func WithResultLimit(h *handler.Server, limit *int) {
	h.AroundFields(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		if limit == nil {
			return next(ctx)
		}

		// grab preloads to set max result limits
		graphutils.GetPreloads(ctx, limit)

		return next(ctx)
	})
}

// WithSkipCache adds a skip cache middleware to the handler
// This is useful for testing, where you don't want to cache responses
// so you can see the changes immediately
func WithSkipCache(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		return next(entcache.Skip(ctx))
	})
}

// WithPool adds a worker pool to the resolver for parallel processing
func (r *Resolver) WithPool(maxWorkers int, includeMetrics bool, options ...pond.Option) {
	// create the pool
	r.pool = soiree.NewPondPool(
		soiree.WithMaxWorkers(maxWorkers),
		soiree.WithName("graphapi-worker-pool"),
		soiree.WithOptions(options...))

	if includeMetrics {
		// add metrics
		r.pool.NewStatsCollector()
	}
}

// Handler returns the http.HandlerFunc for the GraphAPI
func (h *Handler) Handler() http.HandlerFunc {
	return h.graphqlHandler.ServeHTTP
}

// Routes for the the server
func (h *Handler) Routes(e *echo.Group) {
	e.Use(h.middleware...)

	// Create the default POST graph endpoint
	e.POST("/"+graphPath, func(c echo.Context) error {
		h.graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Create a GET query endpoint in order to create short queries with a query string
	e.GET("/"+graphPath, func(c echo.Context) error {
		h.graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
