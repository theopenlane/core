package server

import (
	"context"

	"github.com/rs/zerolog/log"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"

	echo_log "github.com/labstack/gommon/log"

	"github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/route"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/logx/consolelog"
	"github.com/theopenlane/core/pkg/metrics"
	echodebug "github.com/theopenlane/core/pkg/middleware/debug"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// Server is a struct that holds the configuration for the server
type Server struct {
	// config contains the base server settings
	config config.Config
	// handlers contains additional handlers to register with the echo server
	handlers []handler
	// Router makes the router directly accessible on the Server struct
	Router *route.Router
}

// ConfigureEcho sets up the echo server with the default middleware and logging
func ConfigureEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = CustomHTTPErrorHandler
	e.Use(middleware.Recover())

	output := consolelog.NewConsoleWriter()
	logger := logx.New(
		&output,
		logx.WithLevel(echo_log.DEBUG),
		logx.WithTimestamp(),
		logx.WithCaller(),
	)

	e.Logger = logger

	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		TargetHeader: "X-Request-ID",
	}))

	e.Use(logx.LoggingMiddleware(logx.Config{
		Logger:          logger,
		RequestIDHeader: "X-Request-ID",
		RequestIDKey:    "request_id",
		HandleError:     true,
	}))

	return e
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
		Echo: ConfigureEcho(),
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
func NewServer(c config.Config) (*Server, error) {
	srv, err := NewRouter()
	if err != nil {
		return nil, err
	}

	return &Server{
		config: c,
		Router: srv,
	}, nil
}

// StartEchoServer creates and starts the echo server with configured middleware and handlers
func (s *Server) StartEchoServer(ctx context.Context) error {
	sc := echo.StartConfig{
		HideBanner:      true,
		HidePort:        true,
		Address:         s.config.Settings.Server.Listen,
		GracefulTimeout: s.config.Settings.Server.ShutdownGracePeriod,
		GracefulContext: ctx,
	}

	if s.config.Settings.Server.Debug {
		s.Router.Echo.Use(echodebug.BodyDump())
	}

	for _, m := range s.config.DefaultMiddleware {
		s.Router.Echo.Use(m)
	}

	s.Router.Handler = &s.config.Handler

	// Set the local file path if the object storage provider is disk
	// this allows us to serve up the files during testing
	if s.config.Settings.ObjectStorage.Provider == storage.ProviderDisk {
		s.Router.LocalFilePath = s.config.Settings.ObjectStorage.DefaultBucket
	}

	// Add base routes to the server
	if err := route.RegisterRoutes(s.Router); err != nil {
		return err
	}

	// Registers additional routes for the graph endpoints with middleware defined
	for _, handler := range s.handlers {
		handler.Routes(s.Router.Echo.Group("", s.config.GraphMiddleware...))
	}

	// Print routes on startup
	routes := s.Router.Echo.Router().Routes()
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

		return sc.StartTLS(s.Router.Echo, s.config.Settings.Server.TLS.CertFile, s.config.Settings.Server.TLS.CertKey)
	}

	newMetrics := metrics.New(s.config.Settings.Server.MetricsPort)
	go func() {
		if err := newMetrics.Start(ctx); err != nil {
			log.Error().Err(err).Msg("metrics server failed to start")
		}
	}()
	// otherwise, start without TLS
	return sc.Start(s.Router.Echo)
}

var startBlock = `
┌────────────────────────────────────────────────────────────────────────────────────────┐
│      *******   *******  ******** ****     ** **           **     ****     ** ********  │
│     **/////** /**////**/**///// /**/**   /**/**          ****   /**/**   /**/**/////   │
│    **     //**/**   /**/**      /**//**  /**/**         **//**  /**//**  /**/**        │
│   /**      /**/******* /******* /** //** /**/**        **  //** /** //** /**/*******   │
│   /**      /**/**////  /**////  /**  //**/**/**       **********/**  //**/**/**////    │
│   //**     ** /**      /**      /**   //****/**      /**//////**/**   //****/**        │
│    //*******  /**      /********/**    //***/********/**     /**/**    //***/********  │
│     ///////   //       //////// //      /// //////// //      // //      /// ////////   │
└────────────────────────────────────────────────────────────────────────────────────────┘
`
