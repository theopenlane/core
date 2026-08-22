package engine

import "github.com/theopenlane/core/pkg/singleton"

// defaultEngine holds the process-wide workflow engine
var defaultEngine singleton.Value[WorkflowEngine]

// SetDefault registers the process-wide workflow engine
func SetDefault(e *WorkflowEngine) {
	defaultEngine.Set(e)
}

// Default returns the process-wide workflow engine, or nil when none is registered
func Default() *WorkflowEngine {
	return defaultEngine.Get()
}

// Enabled reports whether a process-wide workflow engine is registered
func Enabled() bool {
	return Default() != nil
}
