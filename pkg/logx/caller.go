package logx

import (
	"context"

	"github.com/theopenlane/iam/auth"
)

const (
	// FieldEventID is the log field key for the event envelope identifier
	FieldEventID = "event_id"
	// FieldTopic is the log field key for the event envelope topic
	FieldTopic = "topic"
	// FieldSubjectID is the log field key for the caller subject identifier
	FieldSubjectID = "subject_id"
	// FieldSubjectEmail is the log field key for the caller subject email
	FieldSubjectEmail = "subject_email"
	// FieldOrganizationID is the log field key for the caller organization identifier
	FieldOrganizationID = "organization_id"
	// FieldCapabilities is the log field key for the caller capability bitmask
	FieldCapabilities = "capabilities"
)

// PerHopFields lists the log field keys every durable dispatch hop stamps fresh from its own
// envelope and caller, so they must not ride along in a captured context snapshot
var PerHopFields = []string{
	FieldEventID,
	FieldTopic,
	FieldSubjectID,
	FieldSubjectEmail,
	FieldOrganizationID,
	FieldCapabilities,
}

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
		fields[FieldSubjectID] = caller.SubjectID
	}

	if caller.SubjectEmail != "" {
		fields[FieldSubjectEmail] = caller.SubjectEmail
	}

	if caller.OrganizationID != "" {
		fields[FieldOrganizationID] = caller.OrganizationID
	}

	if caller.Capabilities != 0 {
		fields[FieldCapabilities] = caller.Capabilities
	}

	return WithFields(ctx, fields)
}
