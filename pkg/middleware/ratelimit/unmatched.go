package ratelimit

import (
	"strings"

	echo "github.com/theopenlane/echox"
)

// unmatchedKeyNamespace scopes the unmatched-route limiter key space so it never collides with routed limiter buckets
const unmatchedKeyNamespace = "unmatched"

// UnmatchedRouteLimiterWithConfig returns middleware that rate limits requests which did not match any registered route.
// It is intended to blunt path-enumeration scans (e.g. probing /wp-login.php, /.env) that never reach a real handler.
// The limiter keys solely on the originating IP and never on the request path, because keying on the path would hand a
// scanner a fresh bucket for every unique probe and defeat the limit entirely. Requests that matched a route are passed
// through untouched so the standard routed limiters apply instead
func UnmatchedRouteLimiterWithConfig(conf *Config) echo.MiddlewareFunc {
	if conf == nil {
		conf = &Config{}
	}

	runtime := newRateLimiterRuntime(conf)
	// never advertise Retry-After for unmatched paths: the route does not exist, so a Retry-After would only nudge
	// clients into retrying a request that can never succeed
	runtime.sendRetryAfterHeader = false
	headers := resolveHeaders(conf.Headers, DefaultClientIPHeaders)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// a matched route sets a non-empty path; only unrouted (404) requests are handled here
			if c.Path() != "" {
				return next(c)
			}

			key := buildUnmatchedKey(c, headers, conf)
			if key == "" {
				return next(c)
			}

			if err := runtime.enforce(c, key); err != nil {
				return err
			}

			return next(c)
		}
	}
}

// buildUnmatchedKey derives an IP-only limiter key for unrouted requests, deliberately ignoring path and method
func buildUnmatchedKey(c echo.Context, headers []string, conf *Config) string {
	ip := extractIP(c, headers, conf.ForwardedIndexFromBehind)
	if ip == "" {
		return ""
	}

	parts := []string{}
	if conf.KeyPrefix != "" {
		parts = append(parts, conf.KeyPrefix)
	}

	parts = append(parts, unmatchedKeyNamespace, ip)

	return strings.Join(parts, "|")
}
