package logx

import (
	"context"

	"github.com/rs/zerolog"
)

// WithContext returns a new context with the provided logger
func (l Logger) WithContext(ctx context.Context) context.Context {
	zerologger := l.Unwrap()

	// Check if zerologger is uninitialized by comparing its pointer to nil
	if zerologger.GetLevel() == zerolog.NoLevel {
		return ctx // Return the original context if zerologger is uninitialized
	}

	return zerologger.WithContext(ctx)
}

// Ctx returns a logger from the provided context; if no logger is found in the context, a new one is created
func Ctx(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		defaultLogger := zerolog.New(nil) // Create a default logger
		return &defaultLogger
	}

	return zerolog.Ctx(ctx)
}
