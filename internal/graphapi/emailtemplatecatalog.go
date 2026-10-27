package graphapi

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	access "github.com/theopenlane/core/v2/internal/ent/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	integrationsruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
)

// ensureEmailTemplateViewAccess verifies the caller is a non-anonymous member who can view their active org
func (r *queryResolver) ensureEmailTemplateViewAccess(ctx context.Context) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller.IsAnonymous() || caller.SubjectID == "" {
		return newPermissionDeniedError()
	}

	if caller.Has(auth.CapSystemAdmin) {
		return nil
	}

	orgID, ok := caller.ActiveOrg()
	if !ok {
		return newPermissionDeniedError()
	}

	// email template view access follows org view access through parent_viewer on email_template
	if err := rule.CheckOrgAccess(ctx, fgax.CanView, orgID); err != nil && !access.Allow(err) {
		return newPermissionDeniedError()
	}

	return nil
}

// emailRuntimeClient resolves the runtime email client from the process-wide integrations runtime
func (r *queryResolver) emailRuntimeClient() (*email.Client, error) {
	rt := integrationsruntime.Default()
	if rt == nil {
		return nil, ErrEmailClientNotAvailable
	}

	client, ok := rt.Registry().RuntimeClient(email.DefinitionID.ID())
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	emailClient, ok := client.(*email.Client)
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	return emailClient, nil
}
