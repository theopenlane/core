package anon

import (
	"context"

	"github.com/theopenlane/iam/auth"
)

// TrustCenterScope reports whether ctx carries an anonymous trust center caller
// A caller qualifies only when it has the anonymous role, holds CapTrustCenterAnonymous, and carries
// a non-empty active trust center id in context
func TrustCenterScope(ctx context.Context) (tcID, orgID string, ok bool) {
	caller, callerOK := auth.CallerFromContext(ctx)
	if !callerOK || caller == nil {
		return "", "", false
	}

	if caller.OrganizationRole != auth.AnonymousRole || !caller.Has(auth.CapTrustCenterAnonymous) {
		return "", "", false
	}

	id, hasID := auth.ActiveTrustCenterIDKey.Get(ctx)
	if !hasID || id == "" {
		return "", "", false
	}

	return id, caller.OrganizationID, true
}

// IsTrustCenter reports whether ctx carries an anonymous trust center caller
func IsTrustCenter(ctx context.Context) bool {
	_, _, ok := TrustCenterScope(ctx)

	return ok
}

// IsQuestionnaire reports whether ctx carries an anonymous questionnaire caller
func IsQuestionnaire(ctx context.Context) bool {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return false
	}

	return caller.OrganizationRole == auth.AnonymousRole && caller.Has(auth.CapQuestionnaireAnonymous)
}

// IsAnonymous reports whether ctx carries any anonymous-role caller
func IsAnonymous(ctx context.Context) bool {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return false
	}

	return caller.OrganizationRole == auth.AnonymousRole
}
