package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

// WithInternalCrossOrgContext adds internal operation and the org filter bypass to the current caller:
// skips privacy, FGA, module, and edge checks and reads every org's rows; only for lookups before the org is known
func WithInternalCrossOrgContext(ctx context.Context) context.Context {
	return withCallerCapabilities(ctx, auth.CapInternalOperation|auth.CapBypassOrgFilter)
}

// WithInternalOperationContext adds internal operation to the current caller, keeping its user and orgs:
// skips privacy, FGA, module, and edge checks but reads stay limited to the caller's orgs
func WithInternalOperationContext(ctx context.Context) context.Context {
	return withCallerCapabilities(ctx, auth.CapInternalOperation)
}

// OrgInternalCaller returns a new caller with no user, one org, and internal operation:
// skips privacy, FGA, module, and edge checks, reads only that org's rows, and new rows are owned by that org
func OrgInternalCaller(orgID string) *auth.Caller {
	return &auth.Caller{
		OrganizationID: orgID,
		Capabilities:   auth.CapInternalOperation,
	}
}

// WithOrgInternalCaller replaces the current caller with OrgInternalCaller, dropping its user and any other capabilities
func WithOrgInternalCaller(ctx context.Context, orgID string) context.Context {
	return auth.WithCaller(ctx, OrgInternalCaller(orgID))
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
