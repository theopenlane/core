package logx

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	return FromContext(ctx)
}

// FromContext returns the logger stored on the context or falls back to the global logger.
func FromContext(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return &log.Logger
	}

	logger := zerolog.Ctx(ctx)
	if logger == nil || logger.GetLevel() == zerolog.Disabled {
		return &log.Logger
	}

	return logger
}

// SeedContext ensures the provided context carries a logger, returning a derived context when necessary.
func SeedContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return FromContext(ctx).WithContext(ctx)
}
