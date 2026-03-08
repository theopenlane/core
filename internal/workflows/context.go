package workflows

import (
	"context"

	"entgo.io/ent/privacy"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/utils/contextx"
)

// skipEventEmissionFlag is used to share a mutable skip flag across hook layers.
type skipEventEmissionFlag struct {
	skip bool
}

var skipEventEmissionFlagContextKey = contextx.NewKey[*skipEventEmissionFlag]()

// WithContext sets the workflow bypass flag in the context.
// Operations with this context will skip workflow approval interceptors.
func WithContext(ctx context.Context) context.Context {
	return gala.WithFlag(ctx, gala.ContextFlagWorkflowBypass)
}

// FromContext reports whether the workflow bypass flag is set in the context.
func FromContext(ctx context.Context) bool {
	return gala.HasFlag(ctx, gala.ContextFlagWorkflowBypass)
}

// IsWorkflowBypass checks if the context has workflow bypass enabled.
// Used by workflow interceptors to skip approval routing for system operations.
func IsWorkflowBypass(ctx context.Context) bool {
	return gala.HasFlag(ctx, gala.ContextFlagWorkflowBypass)
}

// WithAllowWorkflowEventEmission marks the context to allow workflow event emission even when bypass is set.
func WithAllowWorkflowEventEmission(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}

	return gala.WithFlag(ctx, gala.ContextFlagWorkflowAllowEventEmission)
}

// AllowWorkflowEventEmission reports whether workflow events should be emitted even when bypass is set.
func AllowWorkflowEventEmission(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	return gala.HasFlag(ctx, gala.ContextFlagWorkflowAllowEventEmission)
}

// WithSkipEventEmission installs a mutable flag in the context so inner hooks can
// signal that mutation events should not be emitted via MarkSkipEventEmission.
func WithSkipEventEmission(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}

	if existing, ok := skipEventEmissionFlagContextKey.Get(ctx); ok && existing != nil {
		return ctx
	}

	return skipEventEmissionFlagContextKey.Set(ctx, &skipEventEmissionFlag{})
}

// MarkSkipEventEmission marks the context to skip emitting mutation events.
func MarkSkipEventEmission(ctx context.Context) {
	if ctx == nil {
		return
	}
	if flag, ok := skipEventEmissionFlagContextKey.Get(ctx); ok && flag != nil {
		flag.skip = true
	}
}

// ShouldSkipEventEmission reports whether mutation event emission should be skipped.
func ShouldSkipEventEmission(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	if flag, ok := skipEventEmissionFlagContextKey.Get(ctx); ok && flag != nil {
		return flag.skip
	}

	return false
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

// AllowBypassContextWithEvents sets workflow bypass, allow decision, and preserves workflow event emission.
func AllowBypassContextWithEvents(ctx context.Context) context.Context {
	return WithAllowWorkflowEventEmission(AllowBypassContext(ctx))
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

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return allowCtx, "", auth.ErrNoAuthUser
	}

	orgID, ok := caller.ActiveOrg()
	if !ok {
		return allowCtx, "", auth.ErrNoAuthUser
	}

	return allowCtx, orgID, nil
}
