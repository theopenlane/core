package ratelimit

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/logx"
)

const (
	defaultRequestsPerWindow = int64(10)
	defaultWindowSize        = time.Second
	defaultFlushInterval     = 10 * time.Second
	defaultRetryMessage      = "Too many requests"
	rateLimitLogComponent    = "ratelimit"
)

// RateOption defines a distinct rate limiting window and request allowance.
type RateOption struct {
	// Requests is the number of requests allowed within the configured window.
	Requests int64 `json:"requests" koanf:"requests" default:"500" example:"500"`
	// Window is the duration of the sliding window.
	Window time.Duration `json:"window" koanf:"window" default:"1m"`
	// Expiration controls how long counters are retained before eviction.
	// When unset, Expiration defaults to twice the Window duration.
	Expiration time.Duration `json:"expiration" koanf:"expiration"`
	// FlushInterval defines how frequently expired values are purged from memory.
	// When unset, FlushInterval defaults to 10s.
	FlushInterval time.Duration `json:"flushinterval" koanf:"flushinterval"`
}

// Config defines the configuration settings for the rate limiter middleware.
type Config struct {
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`
	// Options enables configuring multiple concurrent rate windows that must all pass.
	Options []RateOption `json:"options" koanf:"options"`
	// Headers determines which headers are inspected to determine the origin IP.
	// Defaults to X-Forwarded-For, True-Client-IP, RemoteAddr.
	Headers []string `json:"headers" koanf:"headers" default:"[True-Client-IP]"`
	// ForwardedIndexFromBehind selects which IP from X-Forwarded-For should be used.
	// 0 means the closest client, 1 the proxy behind it, etc.
	ForwardedIndexFromBehind int `json:"forwardedindexfrombehind" koanf:"forwardedindexfrombehind" default:"0"`
	// IncludePath appends the request path to the limiter key when true.
	IncludePath bool `json:"includepath" koanf:"includepath" default:"false"`
	// IncludeMethod appends the request method to the limiter key when true.
	IncludeMethod bool `json:"includemethod" koanf:"includemethod" default:"false"`
	// KeyPrefix allows scoping the limiter key space with a static prefix.
	KeyPrefix string `json:"keyprefix" koanf:"keyprefix"`
	// DenyStatus overrides the HTTP status code returned when a rate limit is exceeded.
	DenyStatus int `json:"denystatus" koanf:"denystatus" default:"429"`
	// DenyMessage customises the error payload when a rate limit is exceeded.
	DenyMessage string `json:"denymessage" koanf:"denymessage" default:"Too many requests"`
	// SendRetryAfterHeader toggles whether the Retry-After header should be added when available.
	SendRetryAfterHeader bool `json:"sendretryafterheader" koanf:"sendretryafterheader" default:"true"`
	// DryRun enables logging rate limit decisions without blocking requests.
	DryRun bool `json:"dryrun" koanf:"dryrun" default:"true"`
}

// RateLimiterWithConfig returns a middleware function for rate limiting requests with a supplied config.
func RateLimiterWithConfig(conf *Config) echo.MiddlewareFunc {
	if conf == nil {
		conf = &Config{}
	}

	limiters := buildLimiters(conf)
	headers := conf.Headers
	if len(headers) == 0 {
		headers = []string{"X-Forwarded-For", "True-Client-IP", "RemoteAddr"}
	}

	denyStatus := conf.DenyStatus
	if denyStatus == 0 {
		denyStatus = http.StatusTooManyRequests
	}

	denyMessage := conf.DenyMessage
	if strings.TrimSpace(denyMessage) == "" {
		denyMessage = defaultRetryMessage
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := buildLimiterKey(c, headers, conf)
			if key == "" {
				return next(c)
			}

			logger := deriveRateLimitLogger(c, key)

			for idx, limiter := range limiters {
				status, err := limiter.Check(key)
				if err != nil {
					logger.
						Error().
						Err(err).
						Int("limiter_index", idx).
						Msg("ratelimit check failed")
					continue
				}

				if status.IsLimited {
					logRateLimitDecision(limitDecisionLogInput{
						Logger:       logger,
						Limiter:      limiter,
						Status:       status,
						LimiterIndex: idx,
						DryRun:       conf.DryRun,
						DenyStatus:   denyStatus,
						DenyMessage:  denyMessage,
					})
					if conf.DryRun {
						continue
					}

					if conf.SendRetryAfterHeader && status.LimitDuration != nil {
						seconds := math.Ceil(status.LimitDuration.Seconds())
						if seconds > 0 {
							c.Response().Header().Set(echo.HeaderRetryAfter, strconv.Itoa(int(seconds)))
						}
					}

					return echo.NewHTTPError(denyStatus, denyMessage)
				}
			}

			for _, limiter := range limiters {
				if err := limiter.Inc(key); err != nil {
					logger.
						Error().
						Err(err).
						Msg("ratelimit increment failed")
				}
			}

			return next(c)
		}
	}
}

func buildLimiters(conf *Config) []*RateLimiter {
	options := conf.Options
	if len(options) == 0 {
		options = []RateOption{{
			Requests: defaultRequestsPerWindow,
			Window:   defaultWindowSize,
		}}
	}

	limiters := make([]*RateLimiter, 0, len(options))

	for _, opt := range options {
		requests := opt.Requests
		if requests <= 0 {
			requests = defaultRequestsPerWindow
		}

		window := opt.Window
		if window <= 0 {
			window = defaultWindowSize
		}

		expiration := opt.Expiration
		if expiration <= 0 {
			expiration = 2 * window // nolint:mnd
		}

		flush := opt.FlushInterval
		if flush <= 0 || flush > expiration {
			flush = defaultFlushInterval
			if flush <= 0 || flush > expiration {
				flush = expiration
			}
		}

		limiters = append(limiters, New(
			NewMapLimitStore(expiration, flush),
			requests,
			window,
		))
	}

	return limiters
}

func buildLimiterKey(c echo.Context, headers []string, conf *Config) string {
	ip := extractIP(c, headers, conf.ForwardedIndexFromBehind)
	if ip == "" {
		return ""
	}

	parts := []string{}
	if conf.KeyPrefix != "" {
		parts = append(parts, conf.KeyPrefix)
	}

	parts = append(parts, ip)

	if conf.IncludeMethod {
		parts = append(parts, strings.ToUpper(c.Request().Method))
	}

	if conf.IncludePath {
		parts = append(parts, c.Path())
	}

	return strings.Join(parts, "|")
}

func deriveRateLimitLogger(c echo.Context, key string) zerolog.Logger {
	base := logx.FromContext(c.Request().Context())

	return base.With().
		Str("component", rateLimitLogComponent).
		Str("ratelimit_key", key).
		Str("request_path", c.Path()).
		Str("request_method", c.Request().Method).
		Str("remote_ip", c.RealIP()).
		Logger()
}

type limitDecisionLogInput struct {
	Logger       zerolog.Logger
	Limiter      *RateLimiter
	Status       *LimitStatus
	LimiterIndex int
	DryRun       bool
	DenyStatus   int
	DenyMessage  string
}

func logRateLimitDecision(input limitDecisionLogInput) {
	event := input.Logger.Warn().
		Int("limiter_index", input.LimiterIndex).
		Bool("dry_run", input.DryRun).
		Float64("current_rate", input.Status.CurrentRate).
		Int64("requests_limit", input.Limiter.RequestsLimit()).
		Dur("window", input.Limiter.WindowSize())

	if input.Status.LimitDuration != nil {
		event.Dur("limit_duration", *input.Status.LimitDuration)
	}

	if !input.DryRun {
		event.
			Int("deny_status", input.DenyStatus).
			Str("deny_message", input.DenyMessage)
	}

	event.Msg("rate limit exceeded")
}

func extractIP(c echo.Context, headers []string, forwardedIndex int) string {
	req := c.Request()

	for _, header := range headers {
		switch strings.ToLower(header) {
		case "remoteaddr":
			if req.RemoteAddr == "" {
				continue
			}

			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				return req.RemoteAddr
			}

			return ip
		case "x-forwarded-for":
			forwarded := req.Header.Get("X-Forwarded-For")
			if forwarded == "" {
				continue
			}

			parts := strings.Split(forwarded, ",")
			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}

			index := max(len(parts)-1-forwardedIndex, 0)

			if parts[index] != "" {
				return parts[index]
			}
		case "True-Client-IP":
			if value := req.Header.Get("True-Client-IP"); value != "" {
				return value
			}
		default:
			if value := req.Header.Get(header); value != "" {
				return value
			}
		}
	}

	return ""
}
