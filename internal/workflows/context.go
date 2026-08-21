package workflows

import (
	"context"

	"entgo.io/ent/privacy"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/pkg/gala"
)

// WithContext sets the workflow bypass flag in the context.
// Operations with this context will skip workflow approval interceptors.
func WithContext(ctx context.Context) context.Context {
	current := gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{})
	current.Bypass = true

	return gala.WorkflowFlagsKey.Set(ctx, current)
}

// IsWorkflowBypass checks if the context has workflow bypass enabled.
// Used by workflow interceptors to skip approval routing for system operations.
func IsWorkflowBypass(ctx context.Context) bool {
	return gala.WorkflowFlagsKey.GetOr(ctx, gala.WorkflowFlags{}).Bypass
}

// withAllowWorkflowEventEmission marks the context to allow workflow event emission even when bypass is set.
func withAllowWorkflowEventEmission(ctx context.Context) context.Context {
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

// AllowContextForOrg returns a privacy-allow context pinned to the supplied organization for
// internal workflow operations. It does not bypass the organization filter; queries stay scoped
// to orgID
func AllowContextForOrg(ctx context.Context, orgID string) context.Context {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	if orgID == "" {
		return allowCtx
	}

	caller, ok := auth.CallerFromContext(allowCtx)
	if !ok || caller == nil {
		caller = &auth.Caller{}
	}

	scoped := *caller
	scoped.OrganizationID = orgID
	scoped.OrganizationIDs = lo.Uniq(append([]string{orgID}, caller.OrgIDs()...))

	return auth.WithCaller(allowCtx, &scoped)
}

// AllowBypassContext sets the workflow-approval bypass flag on a privacy-allow context for
// internal workflow operations
func AllowBypassContext(ctx context.Context) context.Context {
	return WithContext(privacy.DecisionContext(ctx, privacy.Allow))
}

// AllowBypassContextWithEvents sets the workflow-approval bypass flag and preserves workflow event
// emission on a privacy-allow context
func AllowBypassContextWithEvents(ctx context.Context) context.Context {
	return withAllowWorkflowEventEmission(AllowBypassContext(ctx))
}

// AllowContextWithOrg returns a privacy-allow context plus the caller's active organization ID
func AllowContextWithOrg(ctx context.Context) (context.Context, string, error) {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

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
