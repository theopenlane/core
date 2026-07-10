package graphapi

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/definitions/email"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

// emailRuntimeClient resolves the runtime email client from the integrations runtime attached to
// the transactional or resolver db client, matching how campaign dispatch resolves the runtime
func (r *queryResolver) emailRuntimeClient(ctx context.Context) (*email.Client, error) {
	rt := integrationsruntime.FromClient(ctx, r.db)
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
