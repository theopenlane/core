package registry

import "errors"

var (
	// ErrDefinitionIDRequired indicates a definition is missing its canonical identifier
	ErrDefinitionIDRequired = errors.New("integrations/registry: definition id required")
	// ErrDefinitionAlreadyRegistered indicates the definition ID is already present
	ErrDefinitionAlreadyRegistered = errors.New("integrations/registry: definition already registered")
	// ErrDefinitionNotFound indicates the requested definition does not exist
	ErrDefinitionNotFound = errors.New("integrations/registry: definition not found")
	// ErrClientRequired indicates a client registration is missing its identity
	ErrClientRequired = errors.New("integrations/registry: client required")
	// ErrClientAlreadyRegistered indicates a definition already registered the given client name
	ErrClientAlreadyRegistered = errors.New("integrations/registry: client already registered")
	// ErrClientNotFound indicates the requested client does not exist
	ErrClientNotFound = errors.New("integrations/registry: client not found")
	// ErrConnectionCredentialRefRequired indicates a connection registration is missing its credential ref
	ErrConnectionCredentialRefRequired = errors.New("integrations/registry: connection credential ref required")
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
	// ErrOperationHandlerRequired indicates an operation registration is missing both Handle and IngestHandle
	ErrOperationHandlerRequired = errors.New("integrations/registry: operation handler required")
	// ErrOperationHandlerAmbiguous indicates an operation registration specifies both Handle and IngestHandle
	ErrOperationHandlerAmbiguous = errors.New("integrations/registry: operation must specify exactly one of Handle or IngestHandle")
	// ErrIngestContractsRequired indicates an IngestHandle is registered without any Ingest contracts
	ErrIngestContractsRequired = errors.New("integrations/registry: IngestHandle requires at least one Ingest contract")
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
	// ErrOperatorConfigSchemaRequired indicates a definition has an operator config with no schema
	ErrOperatorConfigSchemaRequired = errors.New("integrations/registry: operator config schema required")
	// ErrCredentialSchemaRequired indicates a definition has a credentials block with no schema
	ErrCredentialSchemaRequired = errors.New("integrations/registry: credential schema required")
	// ErrCredentialRefDuplicate indicates a definition registers two credential slots with the same name
	ErrCredentialRefDuplicate = errors.New("integrations/registry: duplicate credential ref name")
	// ErrCredentialRefNotDeclared indicates a client references a credential ref not declared by the definition
	ErrCredentialRefNotDeclared = errors.New("integrations/registry: client credential ref not declared by definition")
	// ErrConnectionRefDuplicate indicates a definition registers two connection modes with the same name
	ErrConnectionRefDuplicate = errors.New("integrations/registry: duplicate connection ref name")
	// ErrConnectionCredentialRefNotDeclared indicates a connection references a credential ref not declared by the definition
	ErrConnectionCredentialRefNotDeclared = errors.New("integrations/registry: connection credential ref not declared by definition")
	// ErrConnectionClientRefNotDeclared indicates a connection references a client ref not declared by the definition
	ErrConnectionClientRefNotDeclared = errors.New("integrations/registry: connection client ref not declared by definition")
	// ErrConnectionValidationOperationNotDeclared indicates a connection validation operation does not exist on the definition
	ErrConnectionValidationOperationNotDeclared = errors.New("integrations/registry: connection validation operation not declared by definition")
	// ErrConnectionAuthCredentialRefNotDeclared indicates a connection auth registration references an undeclared credential ref
	ErrConnectionAuthCredentialRefNotDeclared = errors.New("integrations/registry: connection auth credential ref not declared by connection")
	// ErrConnectionDisconnectCredentialRefNotDeclared indicates a connection disconnect registration references an undeclared credential ref
	ErrConnectionDisconnectCredentialRefNotDeclared = errors.New("integrations/registry: connection disconnect credential ref not declared by connection")
	// ErrUserInputSchemaRequired indicates a definition has a user input block with no schema
	ErrUserInputSchemaRequired = errors.New("integrations/registry: user input schema required")
	// ErrBuilderNil indicates a builder dependency was nil
	ErrBuilderNil = errors.New("integrations/registry: builder is nil")
)
