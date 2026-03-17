package registry

import (
	"cmp"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// Registry is the in-memory index of registered definitions
type Registry struct {
	definitions          map[string]definitionEntry
	operationsByTopic    map[gala.TopicName]types.OperationRegistration
	webhookEventsByTopic map[gala.TopicName]types.WebhookEventRegistration
}

// definitionEntry captures the indexed details for one registered definition
type definitionEntry struct {
	definition    types.Definition
	clients       map[types.ClientID]types.ClientRegistration
	operations    map[string]types.OperationRegistration
	webhooks      map[string]types.WebhookRegistration
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

// Definition returns one definition by canonical identifier
func (r *Registry) Definition(id string) (types.Definition, bool) {
	entry, ok := r.definitions[id]
	if !ok {
		return types.Definition{}, false
	}

	return entry.definition, true
}

// Client returns one client registration for a definition
func (r *Registry) Client(id string, clientID types.ClientID) (types.ClientRegistration, error) {
	return lookupInEntry(r, id, clientID, func(e definitionEntry) map[types.ClientID]types.ClientRegistration {
		return e.clients
	}, ErrClientNotFound)
}

// Operation returns one operation registration for a definition
func (r *Registry) Operation(id string, name string) (types.OperationRegistration, error) {
	return lookupInEntry(r, id, name, func(e definitionEntry) map[string]types.OperationRegistration {
		return e.operations
	}, ErrOperationNotFound)
}

// Webhook returns one webhook registration for a definition
func (r *Registry) Webhook(id string, name string) (types.WebhookRegistration, error) {
	return lookupInEntry(r, id, name, func(e definitionEntry) map[string]types.WebhookRegistration {
		return e.webhooks
	}, ErrWebhookNotFound)
}

// Catalog returns all definition specs in stable id order
func (r *Registry) Catalog() []types.DefinitionSpec {
	out := lo.MapToSlice(r.definitions, func(_ string, entry definitionEntry) types.DefinitionSpec {
		return entry.definition.DefinitionSpec
	})

	slices.SortFunc(out, func(a, b types.DefinitionSpec) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return out
}

// validateDefinition checks the top-level definition identity fields before registration
func (r *Registry) validateDefinition(def types.Definition) error {
	if def.ID == "" {
		return ErrDefinitionIDRequired
	}

	if def.Slug == "" {
		return ErrDefinitionSlugRequired
	}

	if _, exists := r.definitions[def.ID]; exists {
		return ErrDefinitionAlreadyRegistered
	}

	if lo.ContainsBy(lo.Values(r.definitions), func(e definitionEntry) bool {
		return e.definition.Slug == def.Slug
	}) {
		return ErrDefinitionSlugAlreadyRegistered
	}

	return nil
}

// compileDefinition builds the indexed client, operation, and webhook event maps for one definition
func (r *Registry) compileDefinition(def types.Definition) (definitionEntry, error) {
	clients, err := indexClients(def.Clients)
	if err != nil {
		return definitionEntry{}, err
	}

	operations, err := r.indexOperations(def.Operations, clients)
	if err != nil {
		return definitionEntry{}, err
	}

	webhooks, webhookEvents, err := r.indexWebhooks(def.Webhooks)
	if err != nil {
		return definitionEntry{}, err
	}

	return definitionEntry{
		definition:    def,
		clients:       clients,
		operations:    operations,
		webhooks:      webhooks,
		webhookEvents: webhookEvents,
	}, nil
}

// indexClients indexes client registrations by client ref and rejects duplicate or invalid entries
func indexClients(clients []types.ClientRegistration) (map[types.ClientID]types.ClientRegistration, error) {
	clientIndex := make(map[types.ClientID]types.ClientRegistration, len(clients))
	for _, client := range clients {
		if !client.Ref.Valid() {
			return nil, ErrClientRequired
		}

		if _, exists := clientIndex[client.Ref]; exists {
			return nil, ErrClientAlreadyRegistered
		}

		clientIndex[client.Ref] = client
	}

	return clientIndex, nil
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
	out := lo.Values(r.operationsByTopic)

	slices.SortFunc(out, func(a, b types.OperationRegistration) int {
		return cmp.Compare(a.Topic, b.Topic)
	})

	return out
}

// WebhookEvent returns one webhook event registration for a definition
func (r *Registry) WebhookEvent(id string, webhookName string, eventName string) (types.WebhookEventRegistration, error) {
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
	out := lo.Values(r.webhookEventsByTopic)

	slices.SortFunc(out, func(a, b types.WebhookEventRegistration) int {
		return cmp.Compare(a.Topic, b.Topic)
	})

	return out
}

// lookupInEntry finds an entry by definition id, then looks up a value in the sub-map returned by getMap
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
