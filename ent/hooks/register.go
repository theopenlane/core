package hooks

import entgen "github.com/theopenlane/ent/generated"

// RegisterGlobalHooks registers global event hooks for the entdb client and expects a pointer to an Eventer
func RegisterGlobalHooks(client *entgen.Client, e *Eventer) {
	client.Use(EmitEventHook(e))
	client.Use(HookDeletePermissions())
}
