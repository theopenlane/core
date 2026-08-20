package engine

import "sync/atomic"

// defaultEngine holds the process-wide workflow engine so ent hooks and resolvers resolve it
// without a back-edge on the ent client
var defaultEngine atomic.Pointer[WorkflowEngine]

// SetDefault registers the process-wide workflow engine
func SetDefault(e *WorkflowEngine) {
	defaultEngine.Store(e)
}

// Default returns the process-wide workflow engine, or nil when none is registered
func Default() *WorkflowEngine {
	return defaultEngine.Load()
}
