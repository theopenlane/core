package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// rateLimitKeyPrefix namespaces per-operation cooldown keys in redis
const rateLimitKeyPrefix = "integrations:ratelimit:"

// checkRateLimit enforces operation.RateLimit through the shared AllowN window limiter
func (r *Runtime) checkRateLimit(ctx context.Context, operation types.OperationRegistration) (bool, error) {
	if operation.RateLimit == nil {
		return true, nil
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return true, nil
	}

	if caller.Has(auth.CapInternalOperation) {
		logx.FromContext(ctx).Debug().Msg("internal operation caller, bypassing rate limit")

		return true, nil
	}

	orgID, ok := caller.ActiveOrg()
	if !ok {
		return true, nil
	}

	return r.AllowN(ctx, rateLimitKey(operation, orgID), 1, max(operation.RateLimit.Limit, 1), operation.RateLimit.Window)
}

// rateLimitKey builds the key scoping an operation's budget to the calling organization
func rateLimitKey(operation types.OperationRegistration, orgID string) string {
	return fmt.Sprintf("%s:%s", operation.Topic, orgID)
}

// allowNScript atomically increments the window counter, refunding it when the limit is exceeded
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

// AllowN reports whether n more executions fit in limit for window under the namespaced key
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
