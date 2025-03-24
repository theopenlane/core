package echolog

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// WithContext returns a new context with the provided logger
func (l Logger) WithContext(ctx context.Context) context.Context {
	zerologger := l.Unwrap()

	return zerologger.WithContext(ctx)
}

// Ctx returns a logger from the provided context; if no logger is found in the context, a new one is created
func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

func WithNewContext(ctx context.Context, update func(c zerolog.Context) zerolog.Context) context.Context {
	l := log.Ctx(ctx).With().Logger()
	l.UpdateContext(update)

	return l.WithContext(ctx)
}

func WithChildLogger() zerolog.Context {
	return ReturnLogger().With()
}

func ReturnLogger() *zerolog.Logger {
	return &log.Logger
}
