package main

import (
	"context"
	"log"
	"net/http"

	"github.com/zitadel/oidc/example/server/exampleop"
	"github.com/zitadel/oidc/example/server/storage"
)

// Server represents a local OIDC server
type Server struct {
	port string
}

// Option configures the Server
type Option func(*Server)

// WithPort sets the port for the server
func WithPort(p string) Option {
	return func(s *Server) { s.port = p }
}

// New creates a new Server with optional configuration
func New(opts ...Option) *Server {
	s := &Server{port: "9998"}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run starts the server and blocks until the context is done
func (s *Server) Run(ctx context.Context) error {
	store := storage.NewStorage(storage.NewUserStore())
	router := exampleop.SetupServer(ctx, "http://localhost:"+s.port, store)

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: router,
	}

	log.Printf("local OIDC server listening on http://localhost:%s/", s.port)
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

func main() {
	ctx := context.Background()
	srv := New()
	if err := srv.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
