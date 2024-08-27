package server

import (
	"context"

	echo "github.com/theopenlane/echox"
	"go.uber.org/zap"

	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/route"
	echodebug "github.com/theopenlane/core/pkg/middleware/debug"
)

type Server struct {
	// config contains the base server settings
	config config.Config
	// logger contains the zap logger
	logger *zap.SugaredLogger
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
func NewServer(c config.Config, l *zap.SugaredLogger) *Server {
	return &Server{
		config: c,
		logger: l,
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
		srv.Echo.Use(echodebug.BodyDump(s.logger))
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
		s.logger.Infow("registered route", "route", r.Path(), "method", r.Method())
	}

	// if TLS is enabled, start new echo server with TLS
	if s.config.Settings.Server.TLS.Enabled {
		s.logger.Infow("starting in https mode")

		return sc.StartTLS(srv.Echo, s.config.Settings.Server.TLS.CertFile, s.config.Settings.Server.TLS.CertKey)
	}

	s.logger.Infow(startBlock)

	// otherwise, start without TLS
	return sc.Start(srv.Echo)
}

var startBlock = `
┌────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                    │
│                                                                                    │
│                                                                                    │
|        *******                             **                                      │
|       **/////**  ******                   /**                                      │ 
|      **     //**/**///**  *****  *******  /**        ******   *******   *****      │ 
|     /**      /**/**  /** **///**//**///** /**       //////** //**///** **///**     │
|     /**      /**/****** /******* /**  /** /**        *******  /**  /**/*******     │
|     //**     ** /**///  /**////  /**  /** /**       **////**  /**  /**/**////      │
|      //*******  /**     //****** ***  /** /********//******** ***  /**//******     │
|       ///////   //       ////// ///   //  ////////  //////// ///   //  //////      │ 
│                                                                                    │
│                                                                                    │
│                                                                                    │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘`
