package graphapi

import (
	"github.com/theopenlane/core/internal/integrations/definitions/email"
)

// emailRuntimeClient resolves the runtime email client from the integrations runtime
func (r *queryResolver) emailRuntimeClient() (*email.Client, error) {
	client, ok := r.integrationsRuntime.Registry().RuntimeClient(email.DefinitionID.ID())
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	emailClient, ok := client.(*email.Client)
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	return emailClient, nil
}
