package logx

import (
	"context"

	"github.com/theopenlane/iam/auth"
)

// WithCallerIdentity adds the context caller's identity fields to the log context, making
// caller replacement and capability escalation visible on every log line; contexts
// without a caller pass through unchanged
func WithCallerIdentity(ctx context.Context) context.Context {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return ctx
	}

	fields := map[string]any{}

	if caller.SubjectID != "" {
		fields["subject_id"] = caller.SubjectID
	}

	if caller.SubjectEmail != "" {
		fields["subject_email"] = caller.SubjectEmail
	}

	if caller.OrganizationID != "" {
		fields["organization_id"] = caller.OrganizationID
	}

	if caller.Capabilities != 0 {
		fields["capabilities"] = caller.Capabilities
	}

	return WithFields(ctx, fields)
}
