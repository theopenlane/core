package registry

import "errors"

var (
	// ErrDefinitionIDRequired indicates a definition is missing its canonical identifier
	ErrDefinitionIDRequired = errors.New("integrations/registry: definition id required")
	// ErrDefinitionSlugRequired indicates a definition is missing its slug
	ErrDefinitionSlugRequired = errors.New("integrations/registry: definition slug required")
	// ErrDefinitionVersionRequired indicates a definition is missing its version
	ErrDefinitionVersionRequired = errors.New("integrations/registry: definition version required")
	// ErrDefinitionAlreadyRegistered indicates the definition ID is already present
	ErrDefinitionAlreadyRegistered = errors.New("integrations/registry: definition already registered")
	// ErrDefinitionSlugAlreadyRegistered indicates another definition already owns the slug
	ErrDefinitionSlugAlreadyRegistered = errors.New("integrations/registry: definition slug already registered")
	// ErrDefinitionNotFound indicates the requested definition does not exist
	ErrDefinitionNotFound = errors.New("integrations/registry: definition not found")
	// ErrClientRequired indicates a client registration is missing its identity
	ErrClientRequired = errors.New("integrations/registry: client required")
	// ErrClientAlreadyRegistered indicates a definition already registered the given client name
	ErrClientAlreadyRegistered = errors.New("integrations/registry: client already registered")
	// ErrClientNotFound indicates the requested client does not exist
	ErrClientNotFound = errors.New("integrations/registry: client not found")
	// ErrOperationNameRequired indicates an operation registration is missing its name
	ErrOperationNameRequired = errors.New("integrations/registry: operation name required")
	// ErrOperationTopicRequired indicates an operation registration is missing its topic
	ErrOperationTopicRequired = errors.New("integrations/registry: operation topic required")
	// ErrOperationAlreadyRegistered indicates a definition already registered the given operation name
	ErrOperationAlreadyRegistered = errors.New("integrations/registry: operation already registered")
	// ErrOperationTopicAlreadyRegistered indicates another definition already owns the operation topic
	ErrOperationTopicAlreadyRegistered = errors.New("integrations/registry: operation topic already registered")
	// ErrOperationNotFound indicates the requested operation does not exist
	ErrOperationNotFound = errors.New("integrations/registry: operation not found")
	// ErrWebhookNameRequired indicates a webhook registration is missing its name
	ErrWebhookNameRequired = errors.New("integrations/registry: webhook name required")
	// ErrWebhookEventResolverRequired indicates a webhook registration is missing its event resolver
	ErrWebhookEventResolverRequired = errors.New("integrations/registry: webhook event resolver required")
	// ErrWebhookEventHandlerRequired indicates a webhook event registration is missing its handler
	ErrWebhookEventHandlerRequired = errors.New("integrations/registry: webhook event handler required")
	// ErrWebhookAlreadyRegistered indicates a definition already registered the given webhook name
	ErrWebhookAlreadyRegistered = errors.New("integrations/registry: webhook already registered")
	// ErrWebhookNotFound indicates the requested webhook or webhook event does not exist
	ErrWebhookNotFound = errors.New("integrations/registry: webhook not found")
)
