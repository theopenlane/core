package resolvers

import (
	"context"
	"sync"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/workflows"
)

// ResolverFunc resolves a target configuration to a list of user IDs
type ResolverFunc func(ctx context.Context, client *generated.Client, obj *workflows.Object) ([]string, error)

// registry holds all registered resolver functions
// this is intentionally an anonymous struct to limit access and enforce encapsulation
// mutex is used to ensure thread-safe access to the registry
var registry = struct {
	sync.RWMutex
	resolvers map[string]ResolverFunc
}{
	resolvers: make(map[string]ResolverFunc),
}

// Register registers a resolver function with a given key
func Register(key string, fn ResolverFunc) {
	registry.Lock()
	defer registry.Unlock()

	registry.resolvers[key] = fn
}

// Get retrieves a resolver function by key
func Get(key string) (ResolverFunc, bool) {
	registry.RLock()
	defer registry.RUnlock()
	fn, ok := registry.resolvers[key]

	return fn, ok
}

// Keys returns all registered resolver keys
func Keys() []string {
	registry.RLock()
	defer registry.RUnlock()

	return lo.Keys(registry.resolvers)
}
