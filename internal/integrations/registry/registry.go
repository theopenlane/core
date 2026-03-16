package registry

import (
	"sort"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// Registry is the in-memory index of registered definitions
type Registry struct {
	definitionsByID        map[string]types.Definition
	clientsByDefinition    map[string]map[types.ClientID]types.ClientRegistration
	operationsByDefinition map[string]map[string]types.OperationRegistration
	operationsByTopic      map[gala.TopicName]types.OperationRegistration
	webhookEventsByDef     map[string]map[string]map[string]types.WebhookEventRegistration
	webhookEventsByTopic   map[gala.TopicName]types.WebhookEventRegistration
}

// New constructs an empty registry
func New() *Registry {
	return &Registry{
		definitionsByID:        map[string]types.Definition{},
		clientsByDefinition:    map[string]map[types.ClientID]types.ClientRegistration{},
		operationsByDefinition: map[string]map[string]types.OperationRegistration{},
		operationsByTopic:      map[gala.TopicName]types.OperationRegistration{},
		webhookEventsByDef:     map[string]map[string]map[string]types.WebhookEventRegistration{},
		webhookEventsByTopic:   map[gala.TopicName]types.WebhookEventRegistration{},
	}
}

// Register adds one definition to the registry
func (r *Registry) Register(def types.Definition) error {
	definitionID := def.ID
	if definitionID == "" {
		return ErrDefinitionIDRequired
	}

	if def.Slug == "" {
		return ErrDefinitionSlugRequired
	}

	if def.Version == "" {
		return ErrDefinitionVersionRequired
	}

	if _, exists := r.definitionsByID[definitionID]; exists {
		return ErrDefinitionAlreadyRegistered
	}

	for _, existing := range r.definitionsByID {
		if existing.Slug == def.Slug {
			return ErrDefinitionSlugAlreadyRegistered
		}
	}

	clientIndex := make(map[types.ClientID]types.ClientRegistration, len(def.Clients))
	for _, client := range def.Clients {
		if !client.Ref.Valid() {
			return ErrClientRequired
		}

		if _, exists := clientIndex[client.Ref]; exists {
			return ErrClientAlreadyRegistered
		}

		clientIndex[client.Ref] = client
	}

	operationIndex := make(map[string]types.OperationRegistration, len(def.Operations))
	for _, operation := range def.Operations {
		if operation.Name == "" {
			return ErrOperationNameRequired
		}

		if operation.Topic == "" {
			return ErrOperationTopicRequired
		}

		if _, exists := operationIndex[operation.Name]; exists {
			return ErrOperationAlreadyRegistered
		}

		if existing, exists := r.operationsByTopic[operation.Topic]; exists && existing.Name != operation.Name {
			return ErrOperationTopicAlreadyRegistered
		}

		if operation.ClientRef.Valid() {
			if _, exists := clientIndex[operation.ClientRef]; !exists {
				return ErrClientNotFound
			}
		}

		operationIndex[operation.Name] = operation
	}

	webhookNames := make(map[string]struct{}, len(def.Webhooks))
	webhookEventIndex := make(map[string]map[string]types.WebhookEventRegistration, len(def.Webhooks))
	for _, webhook := range def.Webhooks {
		name := webhook.Name
		if name == "" {
			return ErrWebhookNameRequired
		}

		if len(webhook.Events) > 0 && webhook.Event == nil {
			return ErrWebhookEventResolverRequired
		}

		if _, exists := webhookNames[name]; exists {
			return ErrWebhookAlreadyRegistered
		}

		webhookNames[name] = struct{}{}

		eventIndex := make(map[string]types.WebhookEventRegistration, len(webhook.Events))
		for _, event := range webhook.Events {
			if event.Name == "" {
				return ErrWebhookNameRequired
			}

			if event.Topic == "" {
				return ErrOperationTopicRequired
			}

			if event.Handle == nil {
				return ErrWebhookEventHandlerRequired
			}

			if _, exists := eventIndex[event.Name]; exists {
				return ErrWebhookAlreadyRegistered
			}

			if _, exists := r.webhookEventsByTopic[event.Topic]; exists {
				return ErrOperationTopicAlreadyRegistered
			}

			eventIndex[event.Name] = event
		}

		webhookEventIndex[name] = eventIndex
	}

	r.definitionsByID[definitionID] = def
	r.clientsByDefinition[definitionID] = clientIndex
	r.operationsByDefinition[definitionID] = operationIndex
	r.webhookEventsByDef[definitionID] = webhookEventIndex

	for _, operation := range def.Operations {
		r.operationsByTopic[operation.Topic] = operation
	}

	for _, events := range webhookEventIndex {
		for _, event := range events {
			r.webhookEventsByTopic[event.Topic] = event
		}
	}

	return nil
}

// Definition returns one definition by canonical identifier
func (r *Registry) Definition(id string) (types.Definition, bool) {
	definition, ok := r.definitionsByID[id]

	return definition, ok
}

// Client returns one client registration for a definition
func (r *Registry) Client(id string, clientID types.ClientID) (types.ClientRegistration, error) {
	clientIndex, ok := r.clientsByDefinition[id]
	if !ok {
		return types.ClientRegistration{}, ErrDefinitionNotFound
	}

	client, ok := clientIndex[clientID]
	if !ok {
		return types.ClientRegistration{}, ErrClientNotFound
	}

	return client, nil
}

// Operation returns one operation registration for a definition
func (r *Registry) Operation(id string, name string) (types.OperationRegistration, error) {
	operationIndex, ok := r.operationsByDefinition[id]
	if !ok {
		return types.OperationRegistration{}, ErrDefinitionNotFound
	}

	operation, ok := operationIndex[name]
	if !ok {
		return types.OperationRegistration{}, ErrOperationNotFound
	}

	return operation, nil
}

// Catalog returns all definition specs in stable id order
func (r *Registry) Catalog() []types.DefinitionSpec {
	out := make([]types.DefinitionSpec, 0, len(r.definitionsByID))
	for _, definition := range r.definitionsByID {
		out = append(out, definition.DefinitionSpec)
	}

	sort.Slice(out, func(i, j int) bool {
		return string(out[i].ID) < string(out[j].ID)
	})

	return out
}

// Listeners returns all operation registrations in stable topic order
func (r *Registry) Listeners() []types.OperationRegistration {
	out := make([]types.OperationRegistration, 0, len(r.operationsByTopic))
	for _, operation := range r.operationsByTopic {
		out = append(out, operation)
	}

	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Topic) < string(out[j].Topic)
	})

	return out
}

// WebhookEvent returns one webhook event registration for a definition
func (r *Registry) WebhookEvent(id string, webhookName string, eventName string) (types.WebhookEventRegistration, error) {
	webhookIndex, ok := r.webhookEventsByDef[id]
	if !ok {
		return types.WebhookEventRegistration{}, ErrDefinitionNotFound
	}

	eventIndex, ok := webhookIndex[webhookName]
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
	out := make([]types.WebhookEventRegistration, 0, len(r.webhookEventsByTopic))
	for _, event := range r.webhookEventsByTopic {
		out = append(out, event)
	}

	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Topic) < string(out[j].Topic)
	})

	return out
}
