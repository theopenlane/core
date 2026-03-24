package providerkit

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// WithClient adapts a typed client-backed function to a request handler
func WithClient[C, R any](ref types.ClientRef[C], run func(context.Context, C) (R, error)) func(context.Context, types.OperationRequest) (R, error) {
	return WithClientRequest(ref, func(ctx context.Context, _ types.OperationRequest, client C) (R, error) {
		return run(ctx, client)
	})
}

// WithClientRequest adapts a typed client-backed function that also needs the full operation request
func WithClientRequest[C, R any](ref types.ClientRef[C], run func(context.Context, types.OperationRequest, C) (R, error)) func(context.Context, types.OperationRequest) (R, error) {
	return func(ctx context.Context, request types.OperationRequest) (R, error) {
		client, err := ref.Cast(request.Client)
		if err != nil {
			var zero R

			return zero, err
		}

		return run(ctx, request, client)
	}
}

// WithClientConfig adapts a typed client-backed function that also decodes typed config
func WithClientConfig[C, Config, R any](ref types.ClientRef[C], op types.OperationRef[Config], configErr error, run func(context.Context, C, Config) (R, error)) func(context.Context, types.OperationRequest) (R, error) {
	return WithClientRequestConfig(ref, op, configErr, func(ctx context.Context, _ types.OperationRequest, client C, cfg Config) (R, error) {
		return run(ctx, client, cfg)
	})
}

// WithClientRequestConfig adapts a typed client-backed function that needs the full request and typed config
func WithClientRequestConfig[C, Config, R any](ref types.ClientRef[C], op types.OperationRef[Config], configErr error, run func(context.Context, types.OperationRequest, C, Config) (R, error)) func(context.Context, types.OperationRequest) (R, error) {
	return func(ctx context.Context, request types.OperationRequest) (R, error) {
		client, err := ref.Cast(request.Client)
		if err != nil {
			var zero R

			return zero, err
		}

		cfg, err := op.UnmarshalConfig(request.Config)
		if err != nil {
			var zero R

			return zero, configErr
		}

		return run(ctx, request, client, cfg)
	}
}
