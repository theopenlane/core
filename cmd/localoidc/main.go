package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/theopenlane/utils/ulids"
	"github.com/zitadel/oidc/example/server/exampleop"
	"github.com/zitadel/oidc/example/server/storage"
)

// Server represents a local OIDC server.
type Server struct {
	port         string
	redirectURL  string
	clientID     string
	clientSecret string
}

const readHeaderTimeout = 5 * time.Second

// Option configures the Server.
type Option func(*Server)

// WithPort sets the port for the server.
func WithPort(p string) Option {
	return func(s *Server) { s.port = p }
}

// WithRedirectURL sets the redirect URL for the registered client.
func WithRedirectURL(u string) Option {
	return func(s *Server) { s.redirectURL = u }
}

// New creates a new Server with optional configuration.
func New(opts ...Option) *Server {
	s := &Server{
		port:        "9998",
		redirectURL: "http://localhost:17608/v1/sso/callback",
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run starts the server and blocks until the context is done.
func (s *Server) Run(ctx context.Context) error {
	s.clientID = ulids.New().String()
	s.clientSecret = ulids.New().String()

	storage.RegisterClients(storage.WebClient(s.clientID, s.clientSecret, s.redirectURL))
	store := storage.NewStorage(storage.NewUserStore())
	router := exampleop.SetupServer(ctx, "http://localhost:"+s.port, store)

	router.HandleFunc("/client", func(w http.ResponseWriter, _ *http.Request) {
		info := map[string]string{
			"client_id":     s.clientID,
			"client_secret": s.clientSecret,
			"discovery_url": "http://localhost:" + s.port + "/.well-known/openid-configuration",
		}
		_ = json.NewEncoder(w).Encode(info)
	}).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:              ":" + s.port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Printf("local OIDC server listening on http://localhost:%s/", s.port)
	log.Printf("client ID: %s", s.clientID)
	log.Printf("client Secret: %s", s.clientSecret)
	log.Printf("discovery URL: http://localhost:%s/.well-known/openid-configuration", s.port)

	go func() {
		<-ctx.Done()

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	return srv.ListenAndServe()
}

func main() {
	ctx := context.Background()

	opts := []Option{}
	if p := os.Getenv("OIDC_PORT"); p != "" {
		opts = append(opts, WithPort(p))
	}

	if ru := os.Getenv("OIDC_REDIRECT_URL"); ru != "" {
		opts = append(opts, WithRedirectURL(ru))
	}

	srv := New(opts...)
	if err := srv.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
