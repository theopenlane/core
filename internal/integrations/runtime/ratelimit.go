package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// rateLimitKeyPrefix namespaces per-operation cooldown keys in redis
const rateLimitKeyPrefix = "integrations:ratelimit:"

// checkRateLimit enforces operation.RateLimit through the shared AllowN window limiter, consuming one
// execution from the operation's budget for the calling organization. Operations with no RateLimit
// policy, executions with no calling organization, callers holding the internal operation capability,
// or a server with no redis client configured are never limited
func (r *Runtime) checkRateLimit(ctx context.Context, operation types.OperationRegistration) (bool, error) {
	if operation.RateLimit == nil {
		return true, nil
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return true, nil
	}

	if caller.Has(auth.CapInternalOperation) {
		logx.FromContext(ctx).Debug().Str("operation", operation.Name).Msg("internal operation caller, bypassing rate limit")

		return true, nil
	}

	orgID, ok := caller.ActiveOrg()
	if !ok {
		return true, nil
	}

	return r.AllowN(ctx, rateLimitKey(operation, orgID), 1, max(operation.RateLimit.Limit, 1), operation.RateLimit.Window)
}

// rateLimitKey builds the key scoping an operation's execution budget to the calling organization,
// using the operation topic which is unique per definition and operation
func rateLimitKey(operation types.OperationRegistration, orgID string) string {
	return fmt.Sprintf("%s:%s", operation.Topic, orgID)
}

// allowNScript atomically increments the window counter, ensures the key always carries the window
// TTL so redis expiry handles cleanup, and refunds the increment when the limit would be exceeded
// the lua script is a little hard to read but is better than doing multiple round trips with the client methods
var allowNScript = redis.NewScript(`
local count = redis.call('INCRBY', KEYS[1], ARGV[1])
if redis.call('PTTL', KEYS[1]) < 0 then
  redis.call('PEXPIRE', KEYS[1], ARGV[3])
end
if count > tonumber(ARGV[2]) then
  redis.call('DECRBY', KEYS[1], ARGV[1])
  return 0
end
return 1
`)

// AllowN reports whether n additional executions fit within limit per window for the namespaced key,
// consuming n from the window budget when allowed. Keys share the integrations rate limit namespace
// and always expire after window, so no separate cleanup is required. A server with no redis client
// configured is never limited
func (r *Runtime) AllowN(ctx context.Context, key string, n int, limit int, window time.Duration) (bool, error) {
	redisClient := r.Redis()
	if redisClient == nil {
		return true, nil
	}

	namespaced := rateLimitKeyPrefix + key

	allowed, err := allowNScript.Run(ctx, redisClient, []string{namespaced}, n, limit, window.Milliseconds()).Int()
	if err != nil {
		return false, err
	}

	if allowed == 0 {
		logx.FromContext(ctx).Debug().Str("key", namespaced).Int("requested", n).Int("limit", limit).Msg("rate limit window exhausted, denying request")

		return false, nil
	}

	return true, nil
}
