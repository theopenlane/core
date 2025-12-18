package graphapihistory

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/alitto/pond/v2"
	"github.com/gorilla/websocket"
	"github.com/ravilushqa/otelgqlgen"
	"github.com/theopenlane/core/pkg/events/soiree"
	echo "github.com/theopenlane/echox"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/theopenlane/core/internal/ent/historygenerated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/directives"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	gqlhistorygenerated "github.com/theopenlane/core/internal/graphapi/historygenerated"
)

// Resolver provides a graph response resolver
type Resolver struct {
	db                *historygenerated.Client
	pool              *soiree.PondPool
	extensionsEnabled bool
	isDevelopment     bool
	complexityLimit   int
	maxResultLimit    *int
}

// NewResolver returns a resolver configured with the given ent client
func NewResolver(db *historygenerated.Client) *Resolver {
	return &Resolver{
		db: db,
	}
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
	c := &gqlhistorygenerated.Config{Resolvers: r}

	directives.ImplementAllHistoryDirectives(c)

	srv := handler.New(gqlhistorygenerated.NewExecutableSchema(
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

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxMemory: common.DefaultMaxMemoryMB << 20, //nolint:mnd,
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
	common.WithContextLevelCache(srv)

	// add extensions if enabled
	if r.extensionsEnabled {
		common.AddAllExtensions(srv)
	}

	// Set the error presenter to use the custom error presenter
	srv.SetErrorPresenter(gqlerrors.ErrorPresenter)

	// add max result limits to fields in requests
	common.WithResultLimit(srv, r.maxResultLimit)

	common.WithMetrics(srv)

	srv.Use(otelgqlgen.Middleware())

	h := &Handler{
		r:              r,
		graphqlHandler: srv,
	}

	return h
}

func (r *Resolver) WithComplexityLimit(h *handler.Server) {
	// prevent complex queries except the introspection query
	h.Use(common.NewComplexityLimitWithMetrics(func(_ context.Context, rc *graphql.OperationContext) int {
		if rc != nil && rc.OperationName == "IntrospectionQuery" {
			return common.IntrospectionComplexity
		}

		if rc.OperationName == "GlobalSearch" {
			// allow more complexity for the global search
			// e.g. if the complexity limit is 100, we allow 500 for the global search
			return r.complexityLimit * 5 //nolint:mnd
		}

		if r.complexityLimit > 0 {
			return r.complexityLimit
		}

		return common.DefaultComplexityLimit
	}))
}

// WithPool adds a worker pool to the resolver for parallel processing
func (r *Resolver) WithPool(maxWorkers int, includeMetrics bool, options ...pond.Option) {
	// create the pool
	r.pool = soiree.NewPondPool(
		soiree.WithMaxWorkers(maxWorkers),
		soiree.WithName("graphapi-history-worker-pool"),
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
	e.POST("/history/"+common.GraphPath, func(c echo.Context) error {
		h.graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// Create a GET query endpoint in order to create short queries with a query string
	e.GET("/history/"+common.GraphPath, func(c echo.Context) error {
		h.graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
