package echolog

import (
	"context"

	"github.com/rs/zerolog"
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

//func meow() {

//logger := zerolog.New(writer).With().Timestamp().Logger().Hook(
//	zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
//		e.Str("context", logContext)
//	}))
//
//contextHookFunc := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
//	e.Str("user_id", userID)
//})
//}
