package serveropts

import (
	"github.com/rs/zerolog"
)

// LevelNameHook is a hook that sets the level name field to "info" if the level is not set.
type LevelNameHook struct{}

// Run satisfies the zerolog.Hook interface.
func (h LevelNameHook) Run(e *zerolog.Event, l zerolog.Level, _ string) {
	if l == zerolog.NoLevel {
		e.Str(zerolog.LevelFieldName, zerolog.InfoLevel.String())
	}
}
