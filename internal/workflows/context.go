package workflows

import (
	"context"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// WithContext sets the workflow bypass flag in the context.
// Operations with this context will skip workflow approval interceptors.
func WithContext(ctx context.Context) context.Context {
	current := gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{})
	current.Bypass = true

	return gala.WorkflowFlagsKey.Set(ctx, current)
}

// FromContext reports whether the workflow bypass flag is set in the context.
func FromContext(ctx context.Context) bool {
	return gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{}).Bypass
}

// IsWorkflowBypass checks if the context has workflow bypass enabled.
// Used by workflow interceptors to skip approval routing for system operations.
func IsWorkflowBypass(ctx context.Context) bool {
	return gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{}).Bypass
}

// WithAllowWorkflowEventEmission marks the context to allow workflow event emission even when bypass is set.
func WithAllowWorkflowEventEmission(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}

	current := gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{})
	current.AllowEventEmission = true

	return gala.WorkflowFlagsKey.Set(ctx, current)
}

// AllowWorkflowEventEmission reports whether workflow events should be emitted even when bypass is set.
func AllowWorkflowEventEmission(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	return gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{}).AllowEventEmission
}

// AllowContextForOrg returns an allow context scoped to the supplied organization.
func AllowContextForOrg(ctx context.Context, orgID string) context.Context {
	allowCtx := rule.WithInternalContext(ctx)
	if orgID == "" {
		return allowCtx
	}

	caller, ok := auth.CallerFromContext(allowCtx)
	if !ok || caller == nil {
		return allowCtx
	}

	scoped := *caller
	scoped.OrganizationID = orgID
	scoped.OrganizationIDs = lo.Uniq(append([]string{orgID}, caller.OrgIDs()...))

	return auth.WithCaller(allowCtx, &scoped)
}

// AllowBypassContext sets workflow bypass and allow decision for internal workflow operations.
func AllowBypassContext(ctx context.Context) context.Context {
	return WithContext(rule.WithInternalContext(ctx))
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
	allowCtx := rule.WithInternalContext(ctx)
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
