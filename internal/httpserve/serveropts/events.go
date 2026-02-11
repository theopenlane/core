package serveropts

import "github.com/theopenlane/core/pkg/events/soiree"

// WithEventEmitter wires the soiree emitter into handlers
func WithEventEmitter(emitter soiree.Emitter) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if emitter == nil {
			return
		}

		s.Config.Handler.EventEmitter = emitter
	})
}
