package runtime

import (
	"context"
	"fmt"
	"time"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// rateLimitKeyPrefix namespaces per-operation cooldown keys in redis
const rateLimitKeyPrefix = "integrations:ratelimit:"

// checkRateLimit enforces operation.RateLimit for integration, returning false when the owning organization
// has already run this operation within the policy's window. Operations with no RateLimit policy, requests with no resolved integration, or a server with no redis client configured are never limited
func (r *Runtime) checkRateLimit(ctx context.Context, operation types.OperationRegistration, integration *ent.Integration) (bool, error) {
	if operation.RateLimit == nil || integration == nil {
		return true, nil
	}

	redisClient := r.Redis()
	if redisClient == nil {
		return true, nil
	}

	ok, err := redisClient.SetNX(ctx, rateLimitKey(operation, integration), time.Now().Unix(), operation.RateLimit.Window).Result()
	if err != nil {
		return false, err
	}

	return ok, nil
}

// rateLimitKey builds the redis key scoping operation's cooldown to integration's definition and owning organization
func rateLimitKey(operation types.OperationRegistration, integration *ent.Integration) string {
	return fmt.Sprintf("%s%s:%s:%s", rateLimitKeyPrefix, integration.DefinitionID, operation.Name, integration.OwnerID)
}
