package route

import (
	"time"

	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

const (
	// rateLimitWindow is the sliding window applied to every per-route limiter tier
	rateLimitWindow = time.Minute
	// authRateLimitRequests caps requests to sensitive credential endpoints that are brute-force targets
	authRateLimitRequests = int64(10)
	// emailRateLimitRequests caps requests to endpoints that trigger outbound email
	emailRateLimitRequests = int64(5)
	// authFlowRateLimitRequests caps requests to OAuth/SSO/WebAuthn and token-mint flows a legitimate client only hits a handful of times
	authFlowRateLimitRequests = int64(30)
	// publicStaticRateLimitRequests caps requests to public, cheap, cacheable endpoints (static assets, well-known, discovery)
	publicStaticRateLimitRequests = int64(120)
)

// newRouteRateLimit builds a per-route limiter tier for the supplied request budget, keyed on the shared client-IP
// header precedence so per-route keying stays consistent with the global and unmatched limiters behind Cloudflare
func newRouteRateLimit(requests int64) *ratelimit.Config {
	return &ratelimit.Config{
		Enabled:              true,
		Headers:              ratelimit.DefaultClientIPHeaders,
		Options:              []ratelimit.RateOption{{Requests: requests, Window: rateLimitWindow}},
		SendRetryAfterHeader: true,
		DryRun:               true,
	}
}

// authRateLimit is the strict tier for credential endpoints that are common brute-force targets (login, register, TOTP)
var authRateLimit = newRouteRateLimit(authRateLimitRequests)

// emailRateLimit is the tier for endpoints that trigger outbound email, limiting mailbox-flooding abuse
var emailRateLimit = newRouteRateLimit(emailRateLimitRequests)

// authFlowRateLimit is the tier for OAuth/SSO/WebAuthn and token-mint flows that legitimate clients hit only a few times
var authFlowRateLimit = newRouteRateLimit(authFlowRateLimitRequests)

// publicStaticRateLimit is the generous tier for public, cheap, cacheable endpoints (static assets, well-known, discovery)
var publicStaticRateLimit = newRouteRateLimit(publicStaticRateLimitRequests)
