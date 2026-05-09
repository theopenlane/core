package entdb

import "github.com/rs/zerolog/log"

// gooseLogger is a custom logger that implements the goose.Logger interface
type gooseLogger struct {
	enabled bool
}

// newGooseLogger creates a new gooseLogger with the specified enabled state
func newGooseLogger(enabled bool) gooseLogger {
	return gooseLogger{enabled: enabled}
}

// Fatalf logs a fatal error message and exits the application
func (l gooseLogger) Fatalf(format string, v ...any) {
	log.Fatal().Str("component", "goose").Msgf(format, v...)
}

// Printf logs an informational message if logging is enabled
func (l gooseLogger) Printf(format string, v ...any) {
	if !l.enabled {
		return
	}

	log.Info().Str("component", "goose").Msgf(format, v...)
}
