package workflows

import (
	"context"

	"entgo.io/ent/privacy"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// WorkflowBypassContextKey is the context key for workflow bypass operations
// Used to bypass workflow approval checks during system operations (e.g., applying approved changes)
type WorkflowBypassContextKey struct{}

// WithContext sets the workflow bypass context
// Operations with this context will skip workflow approval interceptors
func WithContext(ctx context.Context) context.Context {
	return contextx.With(ctx, WorkflowBypassContextKey{})
}

// FromContext retrieves the workflow bypass context
func FromContext(ctx context.Context) (WorkflowBypassContextKey, bool) {
	return contextx.From[WorkflowBypassContextKey](ctx)
}

// IsWorkflowBypass checks if the context has workflow bypass enabled
// Used by workflow interceptors to skip approval routing for system operations
func IsWorkflowBypass(ctx context.Context) bool {
	_, ok := FromContext(ctx)
	return ok
}

// AllowContext sets the ent privacy decision to allow for internal workflow operations.
// It also sets the internal request marker so FGA checks are bypassed.
func AllowContext(ctx context.Context) context.Context {
	return privacy.DecisionContext(rule.WithInternalContext(ctx), privacy.Allow)
}

// AllowBypassContext sets workflow bypass and allow decision for internal workflow operations.
func AllowBypassContext(ctx context.Context) context.Context {
	return WithContext(AllowContext(ctx))
}

// AllowContextWithOrg returns an allow context plus the organization ID.
func AllowContextWithOrg(ctx context.Context) (context.Context, string, error) {
	return allowContextWithOrg(ctx, false)
}

// AllowBypassContextWithOrg returns an allow/bypass context plus the organization ID.
func AllowBypassContextWithOrg(ctx context.Context) (context.Context, string, error) {
	return allowContextWithOrg(ctx, true)
}

// allowContextWithOrg returns an allow context plus the organization ID with optional workflow bypass
func allowContextWithOrg(ctx context.Context, bypass bool) (context.Context, string, error) {
	allowCtx := AllowContext(ctx)
	if bypass {
		allowCtx = WithContext(allowCtx)
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	return allowCtx, orgID, err
}
