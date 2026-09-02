package hooks

import entgen "github.com/theopenlane/core/v2/internal/ent/generated"

// RegisterGlobalHooks registers global hooks shared across runtime modes.
func RegisterGlobalHooks(client *entgen.Client) {
	client.Use(HookDeletePermissions())
}
