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

// Server represents a local OIDC server
type Server struct {
	Port         string `json:"port"`
	RedirectURL  string `json:"redirect_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

const readHeaderTimeout = 5 * time.Second

// Option configures the Server
type Option func(*Server)

// WithPort sets the Port for the server
func WithPort(p string) Option {
	return func(s *Server) { s.Port = p }
}

// WithRedirectURL sets the redirect URL for the registered client
func WithRedirectURL(u string) Option {
	return func(s *Server) { s.RedirectURL = u }
}

// New creates a new Server with optional configuration
func New(opts ...Option) *Server {
	s := &Server{
		Port:        "9998",
		RedirectURL: "http://localhost:17608/v1/sso/callback",
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Run starts the server and blocks until the context is done.
func (s *Server) Run(ctx context.Context) error {
	s.ClientID = ulids.New().String()
	s.ClientSecret = ulids.New().String()

	storage.RegisterClients(storage.WebClient(s.ClientID, s.ClientSecret, s.RedirectURL))
	store := storage.NewStorage(storage.NewUserStore())
	oidcRouter := exampleop.SetupServer(ctx, "http://localhost:"+s.Port, store)

	mux := http.NewServeMux()
	mux.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		info := map[string]string{
			"port":          s.Port,
			"redirect_url":  s.RedirectURL,
			"client_id":     s.ClientID,
			"client_secret": s.ClientSecret,
			"discovery_url": "http://localhost:" + s.Port + "/.well-known/openid-configuration",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(info)
	})
	mux.Handle("/", oidcRouter)

	srv := &http.Server{
		Addr:              ":" + s.Port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Printf("local OIDC server listening on http://localhost:%s/", s.Port)
	log.Printf("client ID: %s", s.ClientID)
	log.Printf("client Secret: %s", s.ClientSecret)
	log.Printf("discovery URL: http://localhost:%s/.well-known/openid-configuration", s.Port)

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
