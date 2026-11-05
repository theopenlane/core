package graphapi

import (
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

// emailRuntimeClient resolves the runtime email client from the process-wide integrations runtime
func (r *queryResolver) emailRuntimeClient() (*email.Client, error) {
	_, client, ok := runtimeClient(email.DefinitionID.ID())
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	emailClient, ok := client.(*email.Client)
	if !ok {
		return nil, ErrEmailClientNotAvailable
	}

	return emailClient, nil
}
