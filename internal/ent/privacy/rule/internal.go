package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

const internalRequestCaps = auth.CapInternalOperation | auth.CapBypassOrgFilter

// WithInternalCrossOrgContext marks a request as internal by attaching internal bypass
// capabilities to the caller in context.
func WithInternalCrossOrgContext(ctx context.Context) context.Context {
	return withCallerCapabilities(ctx, internalRequestCaps)
}

// WithInternalOperationContext marks a request as internal while keeping the caller's organization filter
func WithInternalOperationContext(ctx context.Context) context.Context {
	return withCallerCapabilities(ctx, auth.CapInternalOperation)
}

// withCallerCapabilities adds the capabilities to the caller in context, creating an empty caller when none is set
func withCallerCapabilities(ctx context.Context, caps auth.Capability) context.Context {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		caller = &auth.Caller{}
	}

	return auth.WithCaller(ctx, caller.WithCapabilities(caps))
}

// IsInternalRequest checks if the context caller has internal operation capability.
func IsInternalRequest(ctx context.Context) bool {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return false
	}

	return caller.Has(auth.CapInternalOperation)
}

// AllowIfInternalRequest is a pre-policy rule that allows all operations if
// the caller carries internal request capability.
func AllowIfInternalRequest() privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		if IsInternalRequest(ctx) {
			return privacy.Allow
		}

		return privacy.Skip
	})
}
