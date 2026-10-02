package registry

import "errors"

var (
	// ErrDefinitionIDRequired indicates a definition is missing its canonical identifier
	ErrDefinitionIDRequired = errors.New("integrations/registry: definition id required")
	// ErrDefinitionAlreadyRegistered indicates the definition ID is already present
	ErrDefinitionAlreadyRegistered = errors.New("integrations/registry: definition already registered")
	// ErrDefinitionNotFound indicates the requested definition does not exist
	ErrDefinitionNotFound = errors.New("integrations/registry: definition not found")
	// ErrDuplicateRegistration indicates a definition registers the same item more than once
	ErrDuplicateRegistration = errors.New("integrations/registry: duplicate registration")
	// ErrClientRequired indicates a client registration is missing its identity
	ErrClientRequired = errors.New("integrations/registry: client required")
	// ErrClientNotFound indicates the requested client does not exist
	ErrClientNotFound = errors.New("integrations/registry: client not found")
	// ErrConnectionCredentialRefRequired indicates a connection is missing its credential ref
	ErrConnectionCredentialRefRequired = errors.New("integrations/registry: connection credential ref required")
	// ErrOperationNotFound indicates the requested operation does not exist
	ErrOperationNotFound = errors.New("integrations/registry: operation not found")
	// ErrOperationHandlerRequired indicates an operation is missing both Handle and IngestHandle
	ErrOperationHandlerRequired = errors.New("integrations/registry: operation handler required")
	// ErrOperationHandlerAmbiguous indicates an operation specifies both Handle and IngestHandle
	ErrOperationHandlerAmbiguous = errors.New("integrations/registry: operation must specify exactly one of Handle or IngestHandle")
	// ErrIngestContractsRequired indicates an IngestHandle is registered without any Ingest contracts
	ErrIngestContractsRequired = errors.New("integrations/registry: IngestHandle requires at least one Ingest contract")
	// ErrIngestSnapshotRequiresIngestHandle indicates Policy.Snapshot is set without an IngestHandle
	ErrIngestSnapshotRequiresIngestHandle = errors.New("integrations/registry: policy snapshot requires an IngestHandle")
	// ErrWebhookEventResolverRequired indicates a webhook registration is missing its event resolver
	ErrWebhookEventResolverRequired = errors.New("integrations/registry: webhook event resolver required")
	// ErrWebhookEventHandlerRequired indicates a webhook event registration is missing its handler
	ErrWebhookEventHandlerRequired = errors.New("integrations/registry: webhook event handler required")
	// ErrWebhookNotFound indicates the requested webhook or webhook event does not exist
	ErrWebhookNotFound = errors.New("integrations/registry: webhook not found")
	// ErrOperatorConfigSchemaRequired indicates a definition has an operator config with no schema
	ErrOperatorConfigSchemaRequired = errors.New("integrations/registry: operator config schema required")
	// ErrCredentialSchemaRequired indicates a credential registration's slot reflects no schema
	ErrCredentialSchemaRequired = errors.New("integrations/registry: credential schema required")
	// ErrCredentialRefNotDeclared indicates a client references an undeclared credential ref
	ErrCredentialRefNotDeclared = errors.New("integrations/registry: client credential ref not declared by definition")
	// ErrConnectionCredentialRefNotDeclared indicates a connection's credential ref is undeclared
	ErrConnectionCredentialRefNotDeclared = errors.New("integrations/registry: connection credential ref not declared by definition")
	// ErrConnectionClientRefNotDeclared indicates a connection's client ref is undeclared
	ErrConnectionClientRefNotDeclared = errors.New("integrations/registry: connection client ref not declared by definition")
	// ErrConnectionHealthCheckHandlerRequired indicates a connection health check has no handler
	ErrConnectionHealthCheckHandlerRequired = errors.New("integrations/registry: connection health check handler required")
	// ErrConnectionAuthCredentialRefNotDeclared indicates auth uses an undeclared credential ref
	ErrConnectionAuthCredentialRefNotDeclared = errors.New("integrations/registry: connection auth credential ref not declared by connection")
	// ErrConnectionDisconnectCredentialRefNotDeclared indicates disconnect uses an undeclared ref
	ErrConnectionDisconnectCredentialRefNotDeclared = errors.New("integrations/registry: connection disconnect credential ref not declared by connection")
	// ErrUserInputSchemaRequired indicates a definition has a user input block with no schema
	ErrUserInputSchemaRequired = errors.New("integrations/registry: user input schema required")
	// ErrBuilderNil indicates a builder dependency was nil
	ErrBuilderNil = errors.New("integrations/registry: builder is nil")
	// ErrRuntimeBuildRequired indicates a runtime integration is missing its Build function
	ErrRuntimeBuildRequired = errors.New("integrations/registry: runtime integration build function required")
	// ErrLinkEdgeNotFound indicates a mapping's link edge does not exist on the source schema
	ErrLinkEdgeNotFound = errors.New("integrations/registry: link edge not found on source schema")
	// ErrLinkEdgeAmbiguous indicates a link rule's target type matches multiple edges, no edge set
	ErrLinkEdgeAmbiguous = errors.New("integrations/registry: link target type is ambiguous, edge name required")
	// ErrLinkRuleInvalid indicates a link rule sets neither or both of a field match and an expression
	ErrLinkRuleInvalid = errors.New("integrations/registry: link rule must set exactly one of field match or expression")
	// ErrLinkTargetNotRegistered indicates a link edge targets a schema without a registry entry
	ErrLinkTargetNotRegistered = errors.New("integrations/registry: link edge target schema is not in the entityops registry")
	// ErrLinkTargetFieldInvalid indicates the target field is not a match key on the target schema
	ErrLinkTargetFieldInvalid = errors.New("integrations/registry: link target field is not a match key on the target schema")
	// ErrLinkSourceFieldInvalid indicates the source field is not a mapped input key of that shape
	ErrLinkSourceFieldInvalid = errors.New("integrations/registry: link source field is not a mapped input key of the required shape")
	// ErrHealthCheckRequired indicates a definition declares connections without a health check
	ErrHealthCheckRequired = errors.New("integrations/registry: health check required when the definition declares connections")
	// ErrHealthCheckClientCredentialMissing indicates the health check misses a credential slot
	ErrHealthCheckClientCredentialMissing = errors.New("integrations/registry: health check client does not use every connection credential slot")
	// ErrOperationFilterExprInvalid indicates a stored operation input carries a filter expression that does not compile
	ErrOperationFilterExprInvalid = errors.New("integrations/registry: operation filter expression invalid")
	// ErrDestructiveSurfaceChange indicates a removed slot, operation, or webhook has no replacement
	ErrDestructiveSurfaceChange = errors.New("integrations/registry: removed credential slot, operation, or webhook has no replacing registration")
)
