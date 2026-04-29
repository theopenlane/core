package providerkit

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
)

// WithClient adapts a typed client-backed function to a request handler with an ugly signature but saves the call sites
func WithClient[C, R any](ref types.ClientRef[C], run func(context.Context, C) (R, error)) func(context.Context, types.OperationRequest) (R, error) {
	return WithClientRequest(ref, func(ctx context.Context, _ types.OperationRequest, client C) (R, error) {
		return run(ctx, client)
	})
}

// WithClientRequest adapts a typed client-backed function + takes the full operation request (also ugly, but good for call sites)
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

// StaticHandler wraps a function that needs no context or request into an OperationHandler
func StaticHandler(run func() (json.RawMessage, error)) types.OperationHandler {
	return func(context.Context, types.OperationRequest) (json.RawMessage, error) {
		return run()
	}
}

// DisabledWhen returns an OperationRegistration.Disabled predicate that unmarshals the
// installation's stored UserInput JSON into T and delegates to the caller-supplied check
func DisabledWhen[T any](check func(T) bool) func(json.RawMessage) bool {
	return func(userInput json.RawMessage) bool {
		var input T
		if err := json.Unmarshal(userInput, &input); err != nil {
			return false
		}

		return check(input)
	}
}

// ConfigFrom returns an OperationRegistration.ConfigResolver that unmarshals the installation's
// stored UserInput JSON into U, extracts the operation-specific config C via fn, and re-encodes
// it as JSON so the ingest pipeline can resolve per-operation filter expressions from it
func ConfigFrom[U any, C any](fn func(U) C) func(json.RawMessage) json.RawMessage {
	return func(userInput json.RawMessage) json.RawMessage {
		var input U
		if err := json.Unmarshal(userInput, &input); err != nil {
			return nil
		}

		out, err := json.Marshal(fn(input))
		if err != nil {
			return nil
		}

		return out
	}
}
