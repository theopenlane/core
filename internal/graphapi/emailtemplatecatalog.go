package graphapi

import (
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

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
