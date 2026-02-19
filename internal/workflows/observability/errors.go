package observability

import "errors"

// ErrEmitNoRuntime is returned when an emit is attempted without a configured gala runtime
var ErrEmitNoRuntime = errors.New("emit requires a gala runtime")
