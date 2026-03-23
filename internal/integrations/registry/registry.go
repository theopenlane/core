package registry

import (
	"sync"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/mapx"
)

// Builder builds one manifest-backed definition
type Builder func() (types.Definition, error)

// Registry is the in-memory index of registered definitions.
// All registrations must complete before concurrent reads begin;
// after that point all query methods are safe for concurrent use
type Registry struct {
	// mu guards all mutable state in the registry
	mu sync.RWMutex
	// definitions maps definition ID to its compiled entry
	definitions map[string]definitionEntry
	// operationsByTopic maps a topic name to its operation registration
	operationsByTopic map[gala.TopicName]types.OperationRegistration
	// webhookEventsByTopic maps a topic name to its webhook event registration
	webhookEventsByTopic map[gala.TopicName]types.WebhookEventRegistration
}

// definitionEntry captures the indexed details for one registered definition
type definitionEntry struct {
	// definition holds the original definition as supplied by the caller
	definition types.Definition
	// connections maps credential ref name to its connection registration
	connections map[string]types.ConnectionRegistration
	// clients maps client ID to its client registration
	clients map[types.ClientID]types.ClientRegistration
	// operations maps operation name to its operation registration
	operations map[string]types.OperationRegistration
	// webhooks maps webhook name to its webhook registration
	webhooks map[string]types.WebhookRegistration
	// webhookEvents maps webhook name to a nested map of event name to event registration
	webhookEvents map[string]map[string]types.WebhookEventRegistration
}

// New constructs an empty registry
func New() *Registry {
	return &Registry{
		definitions:          map[string]definitionEntry{},
		operationsByTopic:    map[gala.TopicName]types.OperationRegistration{},
		webhookEventsByTopic: map[gala.TopicName]types.WebhookEventRegistration{},
	}
}

// Register adds one definition to the registry
func (r *Registry) Register(def types.Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.validateDefinition(def); err != nil {
		return err
	}

	entry, err := r.compileDefinition(def)
	if err != nil {
		return err
	}

	r.definitions[def.ID] = entry
	for _, operation := range entry.operations {
		r.operationsByTopic[operation.Topic] = operation
	}

	for _, events := range entry.webhookEvents {
		for _, event := range events {
			r.webhookEventsByTopic[event.Topic] = event
		}
	}

	return nil
}

// RegisterAll builds and registers every supplied definition builder in order
func (r *Registry) RegisterAll(builders ...Builder) error {
	for _, builder := range builders {
		if builder == nil {
			return ErrBuilderNil
		}

		def, err := builder()
		if err != nil {
			return err
		}

		if err := r.Register(def); err != nil {
			return err
		}
	}

	return nil
}

// Definition returns one definition by canonical identifier
func (r *Registry) Definition(id string) (types.Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.definitions[id]
	if !ok {
		return types.Definition{}, false
	}

	return entry.definition, true
}

// Definitions returns all registered definitions in stable id order
func (r *Registry) Definitions() []types.Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return mapx.SortedProjection(r.definitions, func(e definitionEntry) types.Definition { return e.definition }, func(d types.Definition) string { return d.ID })
}

// Client returns one client registration for a definition
func (r *Registry) Client(id string, clientID types.ClientID) (types.ClientRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lookupInEntry(r, id, clientID, func(e definitionEntry) map[types.ClientID]types.ClientRegistration {
		return e.clients
	}, ErrClientNotFound)
}

// Operation returns one operation registration for a definition
func (r *Registry) Operation(id string, name string) (types.OperationRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lookupInEntry(r, id, name, func(e definitionEntry) map[string]types.OperationRegistration {
		return e.operations
	}, ErrOperationNotFound)
}

// Webhook returns one webhook registration for a definition
func (r *Registry) Webhook(id string, name string) (types.WebhookRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return lookupInEntry(r, id, name, func(e definitionEntry) map[string]types.WebhookRegistration {
		return e.webhooks
	}, ErrWebhookNotFound)
}

// Catalog returns all definition specs in stable id order
func (r *Registry) Catalog() []types.DefinitionSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return mapx.SortedProjection(r.definitions, func(e definitionEntry) types.DefinitionSpec { return e.definition.DefinitionSpec }, func(s types.DefinitionSpec) string { return s.ID })
}

// validateDefinition checks the top-level definition identity fields before registration
func (r *Registry) validateDefinition(def types.Definition) error {
	if def.ID == "" {
		return ErrDefinitionIDRequired
	}

	if _, exists := r.definitions[def.ID]; exists {
		return ErrDefinitionAlreadyRegistered
	}

	if def.OperatorConfig != nil && len(def.OperatorConfig.Schema) == 0 {
		return ErrOperatorConfigSchemaRequired
	}

	authCredentialNames := make(map[string]struct{}, len(def.Connections))
	for _, connection := range def.Connections {
		if connection.Auth != nil && connection.Auth.CredentialRef != (types.CredentialSlotID{}) {
			authCredentialNames[connection.Auth.CredentialRef.String()] = struct{}{}
		}
	}

	if lo.ContainsBy(def.CredentialRegistrations, func(credential types.CredentialRegistration) bool {
		_, authManaged := authCredentialNames[credential.Ref.String()]
		return len(credential.Schema) == 0 && !authManaged
	}) {
		return ErrCredentialSchemaRequired
	}

	if def.UserInput != nil && len(def.UserInput.Schema) == 0 {
		return ErrUserInputSchemaRequired
	}

	return nil
}

// compileDefinition builds the indexed client, operation, and webhook event maps for one definition
func (r *Registry) compileDefinition(def types.Definition) (definitionEntry, error) {
	credentialNames, err := indexCredentialNames(def.CredentialRegistrations)
	if err != nil {
		return definitionEntry{}, err
	}

	clients, err := indexClients(def.Clients, credentialNames)
	if err != nil {
		return definitionEntry{}, err
	}

	operations, err := r.indexOperations(def.Operations, clients)
	if err != nil {
		return definitionEntry{}, err
	}

	connections, err := indexConnections(def.Connections, credentialNames, clients, operations)
	if err != nil {
		return definitionEntry{}, err
	}

	webhooks, webhookEvents, err := r.indexWebhooks(def.Webhooks)
	if err != nil {
		return definitionEntry{}, err
	}

	return definitionEntry{
		definition:    def,
		connections:   connections,
		clients:       clients,
		operations:    operations,
		webhooks:      webhooks,
		webhookEvents: webhookEvents,
	}, nil
}

// indexCredentialNames builds a set of declared credential ref names for a definition,
// returning an error if any two registrations share the same name
func indexCredentialNames(registrations []types.CredentialRegistration) (map[string]struct{}, error) {
	index := make(map[string]struct{}, len(registrations))
	for _, reg := range registrations {
		name := reg.Ref.String()
		if _, exists := index[name]; exists {
			return nil, ErrCredentialRefDuplicate
		}

		index[name] = struct{}{}
	}

	return index, nil
}

// indexClients indexes client registrations by client ref and rejects duplicate or invalid entries;
// credentialNames is the set of declared credential ref names for the definition and every ref listed
// in ClientRegistration.CredentialRefs must appear in this set
func indexClients(clients []types.ClientRegistration, credentialNames map[string]struct{}) (map[types.ClientID]types.ClientRegistration, error) {
	clientIndex := make(map[types.ClientID]types.ClientRegistration, len(clients))
	for _, client := range clients {
		if !client.Ref.Valid() {
			return nil, ErrClientRequired
		}

		if _, exists := clientIndex[client.Ref]; exists {
			return nil, ErrClientAlreadyRegistered
		}

		for _, ref := range client.CredentialRefs {
			if _, declared := credentialNames[ref.String()]; !declared {
				return nil, ErrCredentialRefNotDeclared
			}
		}

		clientIndex[client.Ref] = client
	}

	return clientIndex, nil
}

// credentialRefDeclared reports whether target appears in refs by string comparison
func credentialRefDeclared(refs []types.CredentialSlotID, target types.CredentialSlotID) bool {
	return lo.ContainsBy(refs, func(ref types.CredentialSlotID) bool {
		return ref.String() == target.String()
	})
}

// indexConnections indexes connection registrations while enforcing credential, client, and validation constraints
func indexConnections(connections []types.ConnectionRegistration, credentialNames map[string]struct{}, clients map[types.ClientID]types.ClientRegistration, operations map[string]types.OperationRegistration) (map[string]types.ConnectionRegistration, error) {
	connectionIndex := make(map[string]types.ConnectionRegistration, len(connections))

	for _, connection := range connections {
		if connection.CredentialRef == (types.CredentialSlotID{}) {
			return nil, ErrConnectionCredentialRefRequired
		}

		name := connection.CredentialRef.String()
		if _, exists := connectionIndex[name]; exists {
			return nil, ErrConnectionRefDuplicate
		}

		if _, declared := credentialNames[connection.CredentialRef.String()]; !declared {
			return nil, ErrConnectionCredentialRefNotDeclared
		}

		if !credentialRefDeclared(connection.CredentialRefs, connection.CredentialRef) {
			connection.CredentialRefs = append(connection.CredentialRefs, connection.CredentialRef)
		}

		for _, ref := range connection.CredentialRefs {
			if _, declared := credentialNames[ref.String()]; !declared {
				return nil, ErrConnectionCredentialRefNotDeclared
			}
		}

		for _, clientRef := range connection.ClientRefs {
			if _, declared := clients[clientRef]; !declared {
				return nil, ErrConnectionClientRefNotDeclared
			}
		}

		if connection.ValidationOperation != "" {
			if _, ok := operations[connection.ValidationOperation]; !ok {
				return nil, ErrConnectionValidationOperationNotDeclared
			}
		}

		if connection.Auth != nil {
			if connection.Auth.CredentialRef == (types.CredentialSlotID{}) {
				return nil, ErrConnectionAuthCredentialRefNotDeclared
			}

			if !credentialRefDeclared(connection.CredentialRefs, connection.Auth.CredentialRef) {
				return nil, ErrConnectionAuthCredentialRefNotDeclared
			}
		}

		if connection.Disconnect != nil {
			if connection.Disconnect.CredentialRef == (types.CredentialSlotID{}) {
				return nil, ErrConnectionDisconnectCredentialRefNotDeclared
			}

			if !credentialRefDeclared(connection.CredentialRefs, connection.Disconnect.CredentialRef) {
				return nil, ErrConnectionDisconnectCredentialRefNotDeclared
			}
		}

		connectionIndex[name] = connection
	}

	return connectionIndex, nil
}

// indexOperations indexes operations by name while enforcing topic and client reference constraints
func (r *Registry) indexOperations(operations []types.OperationRegistration, clients map[types.ClientID]types.ClientRegistration) (map[string]types.OperationRegistration, error) {
	operationIndex := make(map[string]types.OperationRegistration, len(operations))
	localTopics := make(map[gala.TopicName]struct{}, len(operations))

	for _, operation := range operations {
		if operation.Name == "" {
			return nil, ErrOperationNameRequired
		}

		if operation.Topic == "" {
			return nil, ErrOperationTopicRequired
		}

		if _, exists := operationIndex[operation.Name]; exists {
			return nil, ErrOperationAlreadyRegistered
		}

		_, local := localTopics[operation.Topic]
		_, global := r.operationsByTopic[operation.Topic]
		if local || global {
			return nil, ErrOperationTopicAlreadyRegistered
		}

		switch {
		case operation.Handle == nil && operation.IngestHandle == nil:
			return nil, ErrOperationHandlerRequired
		case operation.Handle != nil && operation.IngestHandle != nil:
			return nil, ErrOperationHandlerAmbiguous
		case operation.IngestHandle != nil && len(operation.Ingest) == 0:
			return nil, ErrIngestContractsRequired
		}

		if operation.ClientRef.Valid() {
			if _, exists := clients[operation.ClientRef]; !exists {
				return nil, ErrClientNotFound
			}
		}

		localTopics[operation.Topic] = struct{}{}
		operationIndex[operation.Name] = operation
	}

	return operationIndex, nil
}

// indexWebhooks indexes webhook contracts and webhook events while enforcing name and topic uniqueness
func (r *Registry) indexWebhooks(webhooks []types.WebhookRegistration) (map[string]types.WebhookRegistration, map[string]map[string]types.WebhookEventRegistration, error) {
	webhookIndex := make(map[string]types.WebhookRegistration, len(webhooks))
	webhookEventIndex := make(map[string]map[string]types.WebhookEventRegistration, len(webhooks))
	localTopics := make(map[gala.TopicName]struct{})

	for _, webhook := range webhooks {
		name := webhook.Name
		if name == "" {
			return nil, nil, ErrWebhookNameRequired
		}

		if len(webhook.Events) > 0 && webhook.Event == nil {
			return nil, nil, ErrWebhookEventResolverRequired
		}

		if _, exists := webhookIndex[name]; exists {
			return nil, nil, ErrWebhookAlreadyRegistered
		}

		webhookIndex[name] = webhook

		eventIndex := make(map[string]types.WebhookEventRegistration, len(webhook.Events))
		for _, event := range webhook.Events {
			if event.Name == "" {
				return nil, nil, ErrWebhookNameRequired
			}

			if event.Topic == "" {
				return nil, nil, ErrOperationTopicRequired
			}

			if event.Handle == nil {
				return nil, nil, ErrWebhookEventHandlerRequired
			}

			if _, exists := eventIndex[event.Name]; exists {
				return nil, nil, ErrWebhookAlreadyRegistered
			}

			_, local := localTopics[event.Topic]
			_, global := r.webhookEventsByTopic[event.Topic]
			if local || global {
				return nil, nil, ErrOperationTopicAlreadyRegistered
			}

			localTopics[event.Topic] = struct{}{}
			eventIndex[event.Name] = event
		}

		webhookEventIndex[name] = eventIndex
	}

	return webhookIndex, webhookEventIndex, nil
}

// Listeners returns all operation registrations in stable topic order
func (r *Registry) Listeners() []types.OperationRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return mapx.SortedValues(r.operationsByTopic, func(o types.OperationRegistration) gala.TopicName { return o.Topic })
}

// WebhookEvent returns one webhook event registration for a definition
func (r *Registry) WebhookEvent(id string, webhookName string, eventName string) (types.WebhookEventRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.definitions[id]
	if !ok {
		return types.WebhookEventRegistration{}, ErrDefinitionNotFound
	}

	eventIndex, ok := entry.webhookEvents[webhookName]
	if !ok {
		return types.WebhookEventRegistration{}, ErrWebhookNotFound
	}

	event, ok := eventIndex[eventName]
	if !ok {
		return types.WebhookEventRegistration{}, ErrWebhookNotFound
	}

	return event, nil
}

// WebhookListeners returns all webhook event registrations in stable topic order
func (r *Registry) WebhookListeners() []types.WebhookEventRegistration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return mapx.SortedValues(r.webhookEventsByTopic, func(e types.WebhookEventRegistration) gala.TopicName { return e.Topic })
}

// lookupInEntry finds an entry by definition id, then looks up a value in the sub-map returned by getMap;
// callers must hold r.mu.RLock
func lookupInEntry[K comparable, V any](r *Registry, id string, key K, getMap func(definitionEntry) map[K]V, notFoundErr error) (V, error) {
	entry, ok := r.definitions[id]
	if !ok {
		var zero V
		return zero, ErrDefinitionNotFound
	}

	val, ok := getMap(entry)[key]
	if !ok {
		return val, notFoundErr
	}

	return val, nil
}
