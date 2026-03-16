package registry

import (
	"sort"
	"sync"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// Registry is the in-memory index of registered definitions
type Registry struct {
	mu                     sync.RWMutex
	definitionsByID        map[string]types.Definition
	clientsByDefinition    map[string]map[types.ClientID]types.ClientRegistration
	operationsByDefinition map[string]map[string]types.OperationRegistration
	operationsByTopic      map[gala.TopicName]types.OperationRegistration
}

// New constructs an empty registry
func New() *Registry {
	return &Registry{
		definitionsByID:        map[string]types.Definition{},
		clientsByDefinition:    map[string]map[types.ClientID]types.ClientRegistration{},
		operationsByDefinition: map[string]map[string]types.OperationRegistration{},
		operationsByTopic:      map[gala.TopicName]types.OperationRegistration{},
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

	r.mu.Lock()
	defer r.mu.Unlock()

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
	for _, webhook := range def.Webhooks {
		name := webhook.Name
		if name == "" {
			return ErrWebhookNameRequired
		}

		if _, exists := webhookNames[name]; exists {
			return ErrWebhookAlreadyRegistered
		}

		webhookNames[name] = struct{}{}
	}

	r.definitionsByID[definitionID] = def
	r.clientsByDefinition[definitionID] = clientIndex
	r.operationsByDefinition[definitionID] = operationIndex

	for _, operation := range def.Operations {
		r.operationsByTopic[operation.Topic] = operation
	}

	return nil
}

// Definition returns one definition by canonical identifier
func (r *Registry) Definition(id string) (types.Definition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definition, ok := r.definitionsByID[id]

	return definition, ok
}

// Client returns one client registration for a definition
func (r *Registry) Client(id string, clientID types.ClientID) (types.ClientRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

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
	r.mu.RLock()
	defer r.mu.RUnlock()

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
	r.mu.RLock()
	defer r.mu.RUnlock()

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
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]types.OperationRegistration, 0, len(r.operationsByTopic))
	for _, operation := range r.operationsByTopic {
		out = append(out, operation)
	}

	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Topic) < string(out[j].Topic)
	})

	return out
}
