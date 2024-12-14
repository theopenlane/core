package graphapi

import (
	"context"
	"fmt"
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
	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/wundergraph/graphql-go-tools/pkg/playground"

	ent "github.com/theopenlane/core/internal/ent/generated"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/objects"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

const (
	ActionGet    = "get"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionCreate = "create"
)

var (
	graphPath      = "query"
	playgroundPath = "playground"

	graphFullPath = fmt.Sprintf("/%s", graphPath)
)

// Resolver provides a graph response resolver
type Resolver struct {
	db                *ent.Client
	pool              *soiree.PondPool
	extensionsEnabled bool
	uploader          *objects.Objects
}

// NewResolver returns a resolver configured with the given ent client
func NewResolver(db *ent.Client, u *objects.Objects) *Resolver {
	return &Resolver{
		db:       db,
		uploader: u,
	}
}

func (r Resolver) WithExtensions(enabled bool) *Resolver {
	r.extensionsEnabled = enabled

	return &r
}

// Handler is an http handler wrapping a Resolver
type Handler struct {
	r              *Resolver
	graphqlHandler *handler.Server
	playground     *playground.Playground
	middleware     []echo.MiddlewareFunc
}

// Handler returns an http handler for a graph resolver
func (r *Resolver) Handler(withPlayground bool) *Handler {
	srv := handler.New(gqlgenerated.NewExecutableSchema(
		gqlgenerated.Config{
			Resolvers: r,
		},
	))

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second, //nolint:mnd
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		MaxUploadSize: r.uploader.MaxSize,
		MaxMemory:     r.uploader.MaxMemory,
	})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000)) //nolint:mnd

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100), //nolint:mnd
	})

	// add transactional db client
	WithTransactions(srv, r.db)

	// add context level caching
	WithContextLevelCache(srv)

	// add extensions if enabled
	if r.extensionsEnabled {
		AddAllExtensions(srv)
	}

	// add file uploader if it is configured
	if r.uploader != nil {
		WithFileUploader(srv, r.uploader)
	}

	srv.Use(otelgqlgen.Middleware())

	h := &Handler{
		r:              r,
		graphqlHandler: srv,
	}

	if withPlayground {
		h.playground = playground.New(playground.Config{
			PathPrefix:          "/",
			PlaygroundPath:      playgroundPath,
			GraphqlEndpointPath: graphFullPath,
		})
	}

	return h
}

// WithTransactions adds the transactioner to the ent db client
func WithTransactions(h *handler.Server, d *ent.Client) {
	// setup transactional db client
	h.AroundOperations(injectClient(d))

	h.Use(entgql.Transactioner{TxOpener: d})
}

// WithFileUploader adds the file uploader to the graphql handler
// this will handle the file upload process for the multipart form
func WithFileUploader(h *handler.Server, u *objects.Objects) {
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

// WithSkipCache adds a skip cache middleware to the handler
// This is useful for testing, where you don't want to cache responses
// so you can see the changes immediately
func WithSkipCache(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		return next(entcache.Skip(ctx))
	})
}

// WithPool adds a worker pool to the resolver for parallel processing
func (r *Resolver) WithPool(maxWorkers int, options ...pond.Option) {
	// create the pool
	r.pool = soiree.NewPondPool(soiree.WithMaxWorkers(maxWorkers), soiree.WithOptions(options...))
	// add metrics
	r.pool.NewStatsCollector()
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

	if h.playground != nil {
		handlers, err := h.playground.Handlers()
		if err != nil {
			log.Fatal().Err(err).Msg("error configuring playground handlers")

			return
		}

		for i := range handlers {
			// with the function we need to dereference the handler so that it remains
			// the same in the function below
			hCopy := handlers[i].Handler

			e.GET(handlers[i].Path, func(c echo.Context) error {
				hCopy.ServeHTTP(c.Response(), c.Request())

				return nil
			})
		}
	}
}
