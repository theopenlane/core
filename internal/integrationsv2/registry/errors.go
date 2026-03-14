package registry

import "errors"

var (
	// ErrRegistryNil indicates the registry receiver is nil
	ErrRegistryNil = errors.New("integrationsv2/registry: registry is nil")
	// ErrDefinitionIDRequired indicates a definition is missing its canonical identifier
	ErrDefinitionIDRequired = errors.New("integrationsv2/registry: definition id required")
	// ErrDefinitionSlugRequired indicates a definition is missing its slug
	ErrDefinitionSlugRequired = errors.New("integrationsv2/registry: definition slug required")
	// ErrDefinitionVersionRequired indicates a definition is missing its version
	ErrDefinitionVersionRequired = errors.New("integrationsv2/registry: definition version required")
	// ErrDefinitionAlreadyRegistered indicates the definition ID is already present
	ErrDefinitionAlreadyRegistered = errors.New("integrationsv2/registry: definition already registered")
	// ErrDefinitionSlugAlreadyRegistered indicates another definition already owns the slug
	ErrDefinitionSlugAlreadyRegistered = errors.New("integrationsv2/registry: definition slug already registered")
	// ErrDefinitionNotFound indicates the requested definition does not exist
	ErrDefinitionNotFound = errors.New("integrationsv2/registry: definition not found")
	// ErrClientNameRequired indicates a client registration is missing its name
	ErrClientNameRequired = errors.New("integrationsv2/registry: client name required")
	// ErrClientAlreadyRegistered indicates a definition already registered the given client name
	ErrClientAlreadyRegistered = errors.New("integrationsv2/registry: client already registered")
	// ErrClientNotFound indicates the requested client does not exist
	ErrClientNotFound = errors.New("integrationsv2/registry: client not found")
	// ErrOperationNameRequired indicates an operation registration is missing its name
	ErrOperationNameRequired = errors.New("integrationsv2/registry: operation name required")
	// ErrOperationTopicRequired indicates an operation registration is missing its topic
	ErrOperationTopicRequired = errors.New("integrationsv2/registry: operation topic required")
	// ErrOperationAlreadyRegistered indicates a definition already registered the given operation name
	ErrOperationAlreadyRegistered = errors.New("integrationsv2/registry: operation already registered")
	// ErrOperationTopicAlreadyRegistered indicates another definition already owns the operation topic
	ErrOperationTopicAlreadyRegistered = errors.New("integrationsv2/registry: operation topic already registered")
	// ErrOperationNotFound indicates the requested operation does not exist
	ErrOperationNotFound = errors.New("integrationsv2/registry: operation not found")
	// ErrWebhookNameRequired indicates a webhook registration is missing its name
	ErrWebhookNameRequired = errors.New("integrationsv2/registry: webhook name required")
	// ErrWebhookAlreadyRegistered indicates a definition already registered the given webhook name
	ErrWebhookAlreadyRegistered = errors.New("integrationsv2/registry: webhook already registered")
)
