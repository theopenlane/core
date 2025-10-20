package proxy

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

type presignInterceptorBypassFlag bool

const bypassEnabled presignInterceptorBypassFlag = true

// WithPresignInterceptorBypass marks the context so downstream Ent interceptors skip presign handling.
func WithPresignInterceptorBypass(ctx context.Context) context.Context {
	return contextx.With(ctx, bypassEnabled)
}

// ShouldBypassPresignInterceptor reports whether presign interceptors should be skipped.
func ShouldBypassPresignInterceptor(ctx context.Context) bool {
	return contextx.Has[presignInterceptorBypassFlag](ctx)
}
