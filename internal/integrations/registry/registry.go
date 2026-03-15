package registry

import (
	"sort"
	"sync"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// DefinitionRegistry is the read contract consumed by runtime services
type DefinitionRegistry interface {
	// Definition returns one definition by canonical identifier
	Definition(id types.DefinitionID) (types.Definition, bool)
	// Client returns one client registration for a definition
	Client(id types.DefinitionID, name types.ClientName) (types.ClientRegistration, error)
	// Operation returns one operation registration for a definition
	Operation(id types.DefinitionID, name types.OperationName) (types.OperationRegistration, error)
	// OperationFromString returns one operation registration using raw string identifiers
	OperationFromString(definitionID string, name string) (types.OperationRegistration, error)
	// Catalog returns all definition specs in stable id order
	Catalog() []types.DefinitionSpec
	// Listeners returns all operation registrations in stable topic order
	Listeners() []types.OperationRegistration
}

// Registry is the in-memory index of registered definitions
type Registry struct {
	mu                     sync.RWMutex
	definitionsByID        map[types.DefinitionID]types.Definition
	clientsByDefinition    map[types.DefinitionID]map[types.ClientName]types.ClientRegistration
	operationsByDefinition map[types.DefinitionID]map[types.OperationName]types.OperationRegistration
	operationsByTopic      map[gala.TopicName]types.OperationRegistration
	webhooksByDefinition   map[types.DefinitionID][]types.WebhookRegistration
}

type definitionSpecs []types.DefinitionSpec

type operationRegistrations []types.OperationRegistration

// New constructs an empty registry
func New() *Registry {
	return &Registry{
		definitionsByID:        map[types.DefinitionID]types.Definition{},
		clientsByDefinition:    map[types.DefinitionID]map[types.ClientName]types.ClientRegistration{},
		operationsByDefinition: map[types.DefinitionID]map[types.OperationName]types.OperationRegistration{},
		operationsByTopic:      map[gala.TopicName]types.OperationRegistration{},
		webhooksByDefinition:   map[types.DefinitionID][]types.WebhookRegistration{},
	}
}

// Register adds one definition to the registry
func (r *Registry) Register(def types.Definition) error {
	if r == nil {
		return ErrRegistryNil
	}

	definitionID, err := validateDefinition(def)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.definitionsByID[definitionID]; exists {
		return ErrDefinitionAlreadyRegistered
	}

	for _, existing := range r.definitionsByID {
		if existing.Spec.Slug == def.Spec.Slug {
			return ErrDefinitionSlugAlreadyRegistered
		}
	}

	clientIndex := make(map[types.ClientName]types.ClientRegistration, len(def.Clients))
	for _, client := range def.Clients {
		if _, exists := clientIndex[client.Name]; exists {
			return ErrClientAlreadyRegistered
		}

		clientIndex[client.Name] = client
	}

	operationIndex := make(map[types.OperationName]types.OperationRegistration, len(def.Operations))
	for _, operation := range def.Operations {
		if _, exists := operationIndex[operation.Name]; exists {
			return ErrOperationAlreadyRegistered
		}

		if existing, exists := r.operationsByTopic[operation.Topic]; exists && existing.Name != operation.Name {
			return ErrOperationTopicAlreadyRegistered
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

	r.webhooksByDefinition[definitionID] = append([]types.WebhookRegistration(nil), def.Webhooks...)

	return nil
}

// Definition returns one definition by canonical identifier
func (r *Registry) Definition(id types.DefinitionID) (types.Definition, bool) {
	if r == nil {
		return types.Definition{}, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	definition, ok := r.definitionsByID[types.DefinitionID(id)]

	return definition, ok
}

// Client returns one client registration for a definition
func (r *Registry) Client(id types.DefinitionID, name types.ClientName) (types.ClientRegistration, error) {
	if r == nil {
		return types.ClientRegistration{}, ErrRegistryNil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	clientIndex, ok := r.clientsByDefinition[types.DefinitionID(id)]
	if !ok {
		return types.ClientRegistration{}, ErrDefinitionNotFound
	}

	client, ok := clientIndex[name]
	if !ok {
		return types.ClientRegistration{}, ErrClientNotFound
	}

	return client, nil
}

// Operation returns one operation registration for a definition
func (r *Registry) Operation(id types.DefinitionID, name types.OperationName) (types.OperationRegistration, error) {
	if r == nil {
		return types.OperationRegistration{}, ErrRegistryNil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	operationIndex, ok := r.operationsByDefinition[types.DefinitionID(id)]
	if !ok {
		return types.OperationRegistration{}, ErrDefinitionNotFound
	}

	operation, ok := operationIndex[name]
	if !ok {
		return types.OperationRegistration{}, ErrOperationNotFound
	}

	return operation, nil
}

// OperationFromString returns one operation registration for a definition using a raw string identifier
func (r *Registry) OperationFromString(definitionID string, name string) (types.OperationRegistration, error) {
	return r.Operation(types.DefinitionID(definitionID), types.OperationName(name))
}

// Catalog returns all definition specs in stable id order
func (r *Registry) Catalog() []types.DefinitionSpec {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]types.DefinitionSpec, 0, len(r.definitionsByID))
	for _, definition := range r.definitionsByID {
		out = append(out, definition.Spec)
	}

	sort.Sort(definitionSpecs(out))

	return out
}

// Listeners returns all operation registrations in stable topic order
func (r *Registry) Listeners() []types.OperationRegistration {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]types.OperationRegistration, 0, len(r.operationsByTopic))
	for _, operation := range r.operationsByTopic {
		out = append(out, operation)
	}

	sort.Sort(operationRegistrations(out))

	return out
}

// validateDefinition checks the minimal invariants for one definition
func validateDefinition(def types.Definition) (types.DefinitionID, error) {
	definitionID := def.Spec.ID
	if string(definitionID) == "" {
		return "", ErrDefinitionIDRequired
	}

	if def.Spec.Slug == "" {
		return "", ErrDefinitionSlugRequired
	}

	if def.Spec.Version == "" {
		return "", ErrDefinitionVersionRequired
	}

	clientNames := make(map[types.ClientName]struct{}, len(def.Clients))
	for _, client := range def.Clients {
		if client.Name == "" {
			return "", ErrClientNameRequired
		}

		if _, exists := clientNames[client.Name]; exists {
			return "", ErrClientAlreadyRegistered
		}

		clientNames[client.Name] = struct{}{}
	}

	operationNames := make(map[types.OperationName]struct{}, len(def.Operations))
	for _, operation := range def.Operations {
		if operation.Name == "" {
			return "", ErrOperationNameRequired
		}

		if operation.Topic == "" {
			return "", ErrOperationTopicRequired
		}

		if _, exists := operationNames[operation.Name]; exists {
			return "", ErrOperationAlreadyRegistered
		}

		operationNames[operation.Name] = struct{}{}
	}

	return definitionID, nil
}

// Len returns the number of definition specs in the slice
func (s definitionSpecs) Len() int {
	return len(s)
}

// Less reports whether the left definition spec sorts before the right one
func (s definitionSpecs) Less(i int, j int) bool {
	return string(s[i].ID) < string(s[j].ID)
}

// Swap exchanges two definition specs in the slice
func (s definitionSpecs) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

// Len returns the number of operation registrations in the slice
func (s operationRegistrations) Len() int {
	return len(s)
}

// Less reports whether the left operation registration sorts before the right one
func (s operationRegistrations) Less(i int, j int) bool {
	return string(s[i].Topic) < string(s[j].Topic)
}

// Swap exchanges two operation registrations in the slice
func (s operationRegistrations) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}
