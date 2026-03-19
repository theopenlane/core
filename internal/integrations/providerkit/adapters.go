package providerkit

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
)

// OperationWithClient adapts a typed client-backed function to an operation handler.
func OperationWithClient[C any](
	ref types.ClientRef[C],
	run func(context.Context, C) (json.RawMessage, error),
) types.OperationHandler {
	return OperationWithClientRequest(ref, func(ctx context.Context, _ types.OperationRequest, client C) (json.RawMessage, error) {
		return run(ctx, client)
	})
}

// OperationWithClientRequest adapts a typed client-backed function that also needs the full operation request.
func OperationWithClientRequest[C any](
	ref types.ClientRef[C],
	run func(context.Context, types.OperationRequest, C) (json.RawMessage, error),
) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		client, err := ref.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return run(ctx, request, client)
	}
}

// OperationWithClientConfig adapts a typed client-backed function that also decodes typed config.
func OperationWithClientConfig[C any, Config any](
	ref types.ClientRef[C],
	op types.OperationRef[Config],
	configErr error,
	run func(context.Context, C, Config) (json.RawMessage, error),
) types.OperationHandler {
	return OperationWithClientRequestConfig(ref, op, configErr, func(ctx context.Context, _ types.OperationRequest, client C, cfg Config) (json.RawMessage, error) {
		return run(ctx, client, cfg)
	})
}

// OperationWithClientRequestConfig adapts a typed client-backed function that needs the full request and typed config.
func OperationWithClientRequestConfig[C any, Config any](
	ref types.ClientRef[C],
	op types.OperationRef[Config],
	configErr error,
	run func(context.Context, types.OperationRequest, C, Config) (json.RawMessage, error),
) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		client, err := ref.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		cfg, err := op.UnmarshalConfig(request.Config)
		if err != nil {
			if configErr != nil {
				return nil, configErr
			}

			return nil, err
		}

		return run(ctx, request, client, cfg)
	}
}

// IngestWithClientRequest adapts a typed client-backed function that emits ingest payload sets.
func IngestWithClientRequest[C any](
	ref types.ClientRef[C],
	run func(context.Context, types.OperationRequest, C) ([]types.IngestPayloadSet, error),
) types.IngestHandler {
	return func(ctx context.Context, request types.OperationRequest) ([]types.IngestPayloadSet, error) {
		client, err := ref.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		return run(ctx, request, client)
	}
}

// IngestWithClientRequestConfig adapts a typed client-backed function that emits ingest payload sets and decodes typed config.
func IngestWithClientRequestConfig[C any, Config any](
	ref types.ClientRef[C],
	op types.OperationRef[Config],
	configErr error,
	run func(context.Context, types.OperationRequest, C, Config) ([]types.IngestPayloadSet, error),
) types.IngestHandler {
	return func(ctx context.Context, request types.OperationRequest) ([]types.IngestPayloadSet, error) {
		client, err := ref.Cast(request.Client)
		if err != nil {
			return nil, err
		}

		cfg, err := op.UnmarshalConfig(request.Config)
		if err != nil {
			if configErr != nil {
				return nil, configErr
			}

			return nil, err
		}

		return run(ctx, request, client, cfg)
	}
}
