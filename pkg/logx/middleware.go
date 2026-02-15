package logx

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/theopenlane/utils/contextx"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware"
)

// Config defines the config for the echolog middleware
type Config struct {
	// Logger is a custom instance of the logger to use
	Logger *Logger
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper
	// AfterNextSkipper defines a function to skip middleware after the next handler is called
	AfterNextSkipper middleware.Skipper
	// BeforeNext is a function that is executed before the next handler is called
	BeforeNext middleware.BeforeFunc
	// Enricher is a function that can be used to enrich the logger with additional information
	Enricher Enricher
	// RequestIDHeader is the header name to use for the request ID in a log record
	RequestIDHeader string
	// RequestIDKey is the key name to use for the request ID in a log record
	RequestIDKey string
	// NestKey is the key name to use for the nested logger in a log record
	NestKey string
	// HandleError indicates whether to propagate errors up the middleware chain, so the global error handler can decide appropriate status code
	HandleError bool
	// For long-running requests that take longer than this limit, log at a different level
	RequestLatencyLimit time.Duration
	// The level to log at if RequestLatencyLimit is exceeded
	RequestLatencyLevel zerolog.Level
	// AttachRequestMetadata controls whether stable request metadata (client origin IP, user agent, and forwarding headers)
	// is attached to the request-scoped logger context so downstream log entries can include it.
	AttachRequestMetadata bool
}

// Enricher is a function that can be used to enrich the logger with additional information
type Enricher func(c echo.Context, logger zerolog.Context) zerolog.Context

// Context is a wrapper around echo.Context that provides a logger
type Context struct {
	echo.Context
	logger *Logger
}

// NewContext returns a new Context
func NewContext(ctx echo.Context, logger *Logger) *Context {
	return &Context{ctx, logger}
}

// Logger returns the logger from the context
func (c *Context) Logger() echo.Logger {
	return c.logger
}

// LoggingMiddleware is a middleware that logs requests using the provided logger
func LoggingMiddleware(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}

	if config.AfterNextSkipper == nil {
		config.AfterNextSkipper = middleware.DefaultSkipper
	}

	if config.Logger == nil {
		config.Logger = Configure(LoggerConfig{
			Writer:   os.Stdout,
			WithEcho: true,
		}).Echo
	}

	if config.RequestIDKey == "" {
		config.RequestIDKey = "request_id"
	}

	if config.RequestIDHeader == "" {
		config.RequestIDHeader = "X-Request-ID"
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			var err error

			req := c.Request()
			start := time.Now()

			logger := config.Logger

			logger = enrichLogger(c, logger, config)

			// Start with the request context and enrich it with durable fields
			ctx := req.Context()

			id := getRequestID(c, config)
			if id != "" {
				logger = newLoggerFromExisting(logger.log.With().Str(config.RequestIDKey, id).Logger(), logger.out, logger.setters)
				ctx = storeDurableField(ctx, config.RequestIDKey, id)
			}

			if config.AttachRequestMetadata {
				logger, ctx = attachRequestMetadataWithDurableFields(ctx, c, logger)
			}

			// The request context is retrieved and set to the logger's context
			// the context is then set to the request, and a new context is created with the logger
			c.SetRequest(req.WithContext(logger.WithContext(ctx)))
			c = NewContext(c, logger)

			if config.BeforeNext != nil {
				config.BeforeNext(c)
			}

			if err = next(c); err != nil {
				if config.HandleError {
					c.Error(err)
				}
			}

			if config.AfterNextSkipper(c) {
				return nil
			}

			logEvent(c, logger, config, start, err)

			return err
		}
	}
}

// getRequestID retrieves the request ID from the request or response headers
func getRequestID(c echo.Context, config Config) string {
	req := c.Request()
	res := c.Response()

	id := req.Header.Get(config.RequestIDHeader)
	if id == "" {
		id = res.Header().Get(config.RequestIDHeader)
	}

	return id
}

// enrichLogger enriches the logger (lulz) with additional information using the provided Enricher function
func enrichLogger(c echo.Context, logger *Logger, config Config) *Logger {
	if config.Enricher != nil {
		logger = newLoggerFromExisting(logger.log, logger.out, logger.setters)
		logger.log = config.Enricher(c, logger.log.With()).Logger()
	}

	return logger
}

// logEvent logs the event with all the necessary details; it handles errors and latency limits to determine the log level
func logEvent(c echo.Context, logger *Logger, config Config, start time.Time, err error) {
	req := c.Request()
	res := c.Response()
	stop := time.Now()
	latency := stop.Sub(start)

	var mainEvt *zerolog.Event
	// this is the error that's passed in as input from the middleware func

	switch {
	case err != nil:
		mainEvt = logger.log.WithLevel(zerolog.ErrorLevel).Str("error", err.Error())
	case config.RequestLatencyLimit != 0 && latency > config.RequestLatencyLimit:
		mainEvt = logger.log.WithLevel(config.RequestLatencyLevel)
	default:
		mainEvt = logger.log.WithLevel(logger.log.GetLevel())
	}

	var evt *zerolog.Event

	if config.NestKey != "" {
		evt = zerolog.Dict()
	} else {
		evt = mainEvt
	}

	// Only log request metadata in logEvent if it's NOT already in the logger context
	shouldLogRequestMetadata := !config.AttachRequestMetadata || config.NestKey != ""
	if shouldLogRequestMetadata {
		evt.Str("remote_ip", c.RealIP())
		evt.Str("user_agent", req.UserAgent())
		evt.Str("request_protocol", req.Proto)

		if trueClientIP := req.Header.Get("True-Client-IP"); trueClientIP != "" {
			evt.Str("true_client_ip", trueClientIP)
		}

		if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			evt.Str("x_forwarded_for", forwardedFor)
		}

		if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
			evt.Str("x_real_ip", realIP)
		}
	}
	evt.Str("host", req.Host)
	evt.Str("method", req.Method)
	evt.Str("uri", req.RequestURI)
	evt.Int("status", res.Status)
	evt.Str("referer", req.Referer())
	evt.Int64("latency_ms", latency.Milliseconds())
	evt.Str("latency_human", latency.String())

	if query := req.URL.RawQuery; query != "" {
		evt.Str("query", query)
	}

	cl := req.Header.Get(echo.HeaderContentLength)
	if cl == "" {
		cl = "0"
	}

	evt.Str("bytes_in", cl)
	evt.Str("bytes_out", strconv.FormatInt(res.Size, 10))

	if config.NestKey != "" {
		mainEvt.Dict(config.NestKey, evt).Msgf("request details for request to %s %s", req.Method, req.RequestURI)
	} else {
		mainEvt.Msgf("request details for request to %s %s", req.Method, req.RequestURI)
	}
}

// attachRequestMetadataWithDurableFields attaches stable request metadata to the logger and stores fields durably on context.
func attachRequestMetadataWithDurableFields(ctx context.Context, c echo.Context, logger *Logger) (*Logger, context.Context) {
	req := c.Request()
	remoteIP := c.RealIP()

	zctx := logger.log.With().
		Str("remote_ip", remoteIP).
		Str("user_agent", req.UserAgent()).
		Str("request_protocol", req.Proto)

	ctx = storeDurableField(ctx, "remote_ip", remoteIP)
	ctx = storeDurableField(ctx, "user_agent", req.UserAgent())
	ctx = storeDurableField(ctx, "request_protocol", req.Proto)

	if trueClientIP := req.Header.Get("True-Client-IP"); trueClientIP != "" {
		zctx = zctx.Str("true_client_ip", trueClientIP)
		ctx = storeDurableField(ctx, "true_client_ip", trueClientIP)
	}

	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		zctx = zctx.Str("x_forwarded_for", forwardedFor)
		ctx = storeDurableField(ctx, "x_forwarded_for", forwardedFor)
	}

	if realIP := req.Header.Get("X-Real-IP"); realIP != "" {
		zctx = zctx.Str("x_real_ip", realIP)
		ctx = storeDurableField(ctx, "x_real_ip", realIP)
	}

	return newLoggerFromExisting(zctx.Logger(), logger.out, logger.setters), ctx
}

// storeDurableField stores a field in the durable log fields on the context.
func storeDurableField(ctx context.Context, key string, value any) context.Context {
	fields := FieldsFromContext(ctx)
	if fields == nil {
		fields = LogFields{}
	}

	fields[key] = value

	return contextx.With(ctx, fields)
}
