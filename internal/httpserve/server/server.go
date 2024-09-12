package server

import (
	"context"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/route"
	echodebug "github.com/theopenlane/core/pkg/middleware/debug"
)

type Server struct {
	// config contains the base server settings
	config config.Config
	// handlers contains additional handlers to register with the echo server
	handlers []handler
}

type handler interface {
	Routes(*echo.Group)
}

// NewRouter creates a wrapper router so that the echo server and OAS specification can be generated simultaneously
func NewRouter() (*route.Router, error) {
	oas, err := NewOpenAPISpec()
	if err != nil {
		return nil, err
	}

	return &route.Router{
		Echo: echo.New(),
		OAS:  oas,
	}, nil
}

// AddHandler provides the ability to add additional HTTP handlers that process
// requests. The handler that is provided should have a Routes(*echo.Group)
// function, which allows the routes to be added to the server.
func (s *Server) AddHandler(r handler) {
	s.handlers = append(s.handlers, r)
}

// NewServer returns a new Server configuration
func NewServer(c config.Config) *Server {
	return &Server{
		config: c,
	}
}

// StartEchoServer creates and starts the echo server with configured middleware and handlers
func (s *Server) StartEchoServer(ctx context.Context) error {
	srv, err := NewRouter()
	if err != nil {
		return err
	}

	sc := echo.StartConfig{
		HideBanner:      true,
		HidePort:        true,
		Address:         s.config.Settings.Server.Listen,
		GracefulTimeout: s.config.Settings.Server.ShutdownGracePeriod,
		GracefulContext: ctx,
	}

	srv.Echo.Debug = s.config.Settings.Server.Debug

	if srv.Echo.Debug {
		srv.Echo.Use(echodebug.BodyDump())
	}

	for _, m := range s.config.DefaultMiddleware {
		srv.Echo.Use(m)
	}

	srv.Handler = &s.config.Handler

	// Add base routes to the server
	if err := route.RegisterRoutes(srv); err != nil {
		return err
	}

	// Registers additional routes for the graph endpoints with middleware defined
	for _, handler := range s.handlers {
		handler.Routes(srv.Echo.Group("", s.config.GraphMiddleware...))
	}

	// Print routes on startup
	routes := srv.Echo.Router().Routes()
	for _, r := range routes {
		log.Info().
			Str("route", r.Path()).
			Str("method", r.Method()).
			Msg("registered route")
	}

	log.Info().Msg(startBlock)

	// if TLS is enabled, start new echo server with TLS
	if s.config.Settings.Server.TLS.Enabled {
		log.Info().Msg("starting in https mode")

		return sc.StartTLS(srv.Echo, s.config.Settings.Server.TLS.CertFile, s.config.Settings.Server.TLS.CertKey)
	}

	// otherwise, start without TLS
	return sc.Start(srv.Echo)
}

var startBlock = `
________________________________________________________________
                                                                
                                     /                          
-----------__------__----__----__---/----__----__----__---------
         /   )   /   ) /___) /   ) /   /   ) /   ) /___)        
________(___/___/___/_(___ _/___/_/___(___(_/___/_(___ _________     

`
