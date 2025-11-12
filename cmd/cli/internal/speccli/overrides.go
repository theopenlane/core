//go:build cli

package speccli

import (
	"strings"
	"sync"
)

var (
	overrideMu       sync.RWMutex
	overrideRegistry = map[string]SpecOverride{}
)

// RegisterOverride registers a mutation function that runs after a spec is loaded.
// Passing a nil override removes any previously registered override for the key.
func RegisterOverride(name string, override SpecOverride) {
	key := normalizeOverrideKey(name)

	overrideMu.Lock()
	defer overrideMu.Unlock()

	if override == nil {
		delete(overrideRegistry, key)
		return
	}

	overrideRegistry[key] = override
}

// lookupOverride returns the override registered under the normalized key.
func lookupOverride(name string) (SpecOverride, bool) {
	key := normalizeOverrideKey(name)

	overrideMu.RLock()
	defer overrideMu.RUnlock()

	override, ok := overrideRegistry[key]
	return override, ok
}

// normalizeOverrideKey standardizes override identifiers for case-insensitive lookups.
func normalizeOverrideKey(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}
