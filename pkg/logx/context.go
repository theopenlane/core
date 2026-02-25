package logx

import (
	"context"
	"maps"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/utils/contextx"
)

// LogFields holds structured log fields that can be captured and restored.
type LogFields map[string]any

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

// WithField adds a single field to both the logger and the durable field store on the context.
// A new LogFields map is allocated on each call so that sibling contexts do not share mutable state.
func WithField(ctx context.Context, key string, value any) context.Context {
	existing := FieldsFromContext(ctx)
	fields := make(LogFields, len(existing)+1)

	maps.Copy(fields, existing)

	fields[key] = value

	logger := FromContext(ctx).With().Interface(key, value).Logger()

	ctx = contextx.With(ctx, fields)

	return logger.WithContext(ctx)
}

// WithFields adds multiple fields to both the logger and the durable field store on the context.
// A new LogFields map is allocated on each call so that sibling contexts do not share mutable state.
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	if len(fields) == 0 {
		return ctx
	}

	existing := FieldsFromContext(ctx)
	merged := make(LogFields, len(existing)+len(fields))

	for k, v := range existing {
		merged[k] = v
	}

	logCtx := FromContext(ctx).With()

	for k, v := range fields {
		merged[k] = v
		logCtx = logCtx.Interface(k, v)
	}

	ctx = contextx.With(ctx, merged)

	return logCtx.Logger().WithContext(ctx)
}

// FieldsFromContext returns the durable log fields stored on the context.
func FieldsFromContext(ctx context.Context) LogFields {
	if ctx == nil {
		return nil
	}

	fields, _ := contextx.From[LogFields](ctx)

	return fields
}
