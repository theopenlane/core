package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/mapx"
)

// Builder builds one manifest-backed definition
type Builder func() (types.Definition, error)

// Registry is the in-memory index of registered definitions.
// Built once at startup via RegisterAll; all state is read-only after construction
type Registry struct {
	// definitions maps definition ID to its compiled entry
	definitions map[string]definitionEntry
	// operationsByTopic maps a topic name to its operation registration
	operationsByTopic map[gala.TopicName]types.OperationRegistration
	// webhookEventsByTopic maps a topic name to its webhook event registration
	webhookEventsByTopic map[gala.TopicName]types.WebhookEventRegistration
	// galaListeners collects standalone gala listener registrations across definitions
	galaListeners []types.GalaListenerRegistration
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
	// runtimeClient holds the pre-built client for runtime integrations.
	// Non-nil only when the definition has a RuntimeIntegration with populated config
	runtimeClient any
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

	populateMappingLinkTargets(def.Mappings)

	entry, err := compileDefinition(def)
	if err != nil {
		return err
	}

	if def.RuntimeIntegration != nil && def.RuntimeIntegration.Config != nil {
		client, buildErr := def.RuntimeIntegration.Build(context.Background(), def.RuntimeIntegration.Config)
		if buildErr != nil {
			return buildErr
		}

		entry.runtimeClient = client
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

	r.galaListeners = append(r.galaListeners, def.GalaListeners...)

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
	entry, ok := r.definitions[id]
	if !ok {
		return types.Definition{}, false
	}

	return entry.definition, true
}

// Definitions returns all registered definitions in stable id order
func (r *Registry) Definitions() []types.Definition {
	return mapx.SortedProjection(r.definitions, func(e definitionEntry) types.Definition { return e.definition }, func(d types.Definition) string { return d.ID })
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

	if def.RuntimeIntegration != nil {
		if def.RuntimeIntegration.Build == nil {
			return ErrRuntimeBuildRequired
		}
	}

	if err := validateMappingLinks(def.Mappings); err != nil {
		return err
	}

	return nil
}

// populateMappingLinkTargets fills each mapping's cross-link inventory from the entityops catalog —
// one entry per edge, carrying the edge name, the target's match-key fields, and the source's mapped
// input keys — so the definition payload surfaces the exact identifiers a LinkRule may reference and
// configuration never falls back to free-typed field names
func populateMappingLinkTargets(mappings []types.MappingRegistration) {
	for i := range mappings {
		sourceSchema, ok := entityops.LookupSchema(mappings[i].Schema)
		if !ok {
			continue
		}

		sourceFields := sourceLinkFields(sourceSchema.Fields)

		var targets []types.LinkTargetInfo

		for _, edge := range sourceSchema.Edges {
			if edge.TargetType == "" || edge.Target == nil {
				continue
			}

			targets = append(targets, types.LinkTargetInfo{
				Edge:         edge.Name,
				TargetType:   edge.TargetType,
				Label:        edge.Label,
				TargetFields: targetLinkFields(edge.Target.Fields),
				SourceFields: sourceFields,
			})
		}

		mappings[i].LinkTargets = targets
	}
}

// targetLinkFields projects the match-key (indexed string) fields of a target schema — the fields a
// LinkRule.TargetField may name
func targetLinkFields(fields []entityops.FieldDescriptor) []types.LinkFieldInfo {
	return lo.FilterMap(fields, func(f entityops.FieldDescriptor, _ int) (types.LinkFieldInfo, bool) {
		if !f.MatchKey {
			return types.LinkFieldInfo{}, false
		}

		return types.LinkFieldInfo{Name: f.Name, Label: f.Label, Type: f.Type}, true
	})
}

// sourceLinkFields projects the mapped input keys of the source schema — the keys present in the
// mapped ingest payload that a LinkRule.SourceField (scalar) or SourceList (list) may name
func sourceLinkFields(fields []entityops.FieldDescriptor) []types.LinkFieldInfo {
	return lo.FilterMap(fields, func(f entityops.FieldDescriptor, _ int) (types.LinkFieldInfo, bool) {
		if f.InputKey == "" {
			return types.LinkFieldInfo{}, false
		}

		return types.LinkFieldInfo{Name: f.InputKey, Label: f.Label, Type: f.Type}, true
	})
}

// ResolveLinkEdge resolves the edge a link rule links through: an explicit rule edge is looked up by
// name and checked against the declared target type; otherwise the target type must identify exactly
// one edge, since silently picking one of several (e.g. editors vs viewers, both targeting Group)
// would link through an arbitrary edge
func ResolveLinkEdge(sourceSchema *entityops.Schema, rule types.LinkRule) (entityops.EdgeDescriptor, error) {
	if rule.Edge != "" {
		edge, found := sourceSchema.EdgeByName(rule.Edge)
		if !found {
			return entityops.EdgeDescriptor{}, fmt.Errorf("%w: %s has no edge %s", ErrLinkEdgeNotFound, sourceSchema.Name, rule.Edge)
		}

		if rule.TargetSchema != "" && edge.TargetType != rule.TargetSchema {
			return entityops.EdgeDescriptor{}, fmt.Errorf("%w: edge %s.%s targets %s, not %s", ErrLinkEdgeNotFound, sourceSchema.Name, rule.Edge, edge.TargetType, rule.TargetSchema)
		}

		return edge, nil
	}

	candidates := lo.Filter(sourceSchema.Edges, func(e entityops.EdgeDescriptor, _ int) bool {
		return e.TargetType == rule.TargetSchema
	})

	switch len(candidates) {
	case 0:
		return entityops.EdgeDescriptor{}, fmt.Errorf("%w: %s has no edge to %s", ErrLinkEdgeNotFound, sourceSchema.Name, rule.TargetSchema)
	case 1:
		return candidates[0], nil
	default:
		names := lo.Map(candidates, func(e entityops.EdgeDescriptor, _ int) string { return e.Name })

		return entityops.EdgeDescriptor{}, fmt.Errorf("%w: %s has %d edges to %s (%s); set the rule's edge", ErrLinkEdgeAmbiguous, sourceSchema.Name, len(candidates), rule.TargetSchema, strings.Join(names, ", "))
	}
}

// validateMappingLinks verifies every link rule a mapping declares against the entityops catalog —
// the edge resolves unambiguously, the match shape is coherent, the target field is a match key on
// the target, and the source fields are mapped input keys of the right shape — so a typo or an
// ambiguous target in a definition's link defaults fails at registration instead of silently
// misbehaving at ingest
func validateMappingLinks(mappings []types.MappingRegistration) error {
	for _, mapping := range mappings {
		if len(mapping.Spec.Links) == 0 {
			continue
		}

		sourceSchema, ok := entityops.LookupSchema(mapping.Schema)
		if !ok {
			continue
		}

		if err := ValidateLinkRules(sourceSchema, mapping.Spec.Links); err != nil {
			return err
		}
	}

	return nil
}

// ValidateLinkRules validates each rule against the source schema's catalog: the edge resolves
// unambiguously, the match shape is coherent, and the referenced fields exist with the right
// shape. It is shared by definition registration and installation config validation so a bad
// rule fails at declaration or save time rather than at ingest
func ValidateLinkRules(sourceSchema *entityops.Schema, rules []types.LinkRule) error {
	for _, rule := range rules {
		edge, err := ResolveLinkEdge(sourceSchema, rule)
		if err != nil {
			return err
		}

		if err := validateLinkRuleFields(sourceSchema, edge, rule); err != nil {
			return err
		}
	}

	return nil
}

// validateLinkRuleFields checks one resolved rule's match configuration against the catalogs of the
// source and target schemas
func validateLinkRuleFields(sourceSchema *entityops.Schema, edge entityops.EdgeDescriptor, rule types.LinkRule) error {
	fieldMatch := rule.TargetField != "" && (rule.SourceField != "" || rule.SourceList != "")

	if fieldMatch == (rule.Expression != "") {
		return fmt.Errorf("%w: %s -> %s must set either a target/source field match or an expression", ErrLinkRuleInvalid, sourceSchema.Name, edge.Name)
	}

	if !fieldMatch {
		return nil
	}

	if edge.Target == nil || !edge.Target.MatchKeyField(rule.TargetField) {
		return fmt.Errorf("%w: %s is not a match-key field on %s", ErrLinkTargetFieldInvalid, rule.TargetField, edge.TargetType)
	}

	if rule.SourceField != "" {
		if err := validateSourceKey(sourceSchema, rule.SourceField, false); err != nil {
			return err
		}
	}

	if rule.SourceList != "" {
		if err := validateSourceKey(sourceSchema, rule.SourceList, true); err != nil {
			return err
		}
	}

	return nil
}

// validateSourceKey checks that key is a mapped input key on the source schema whose type shape
// (scalar vs list) matches its LinkRule slot
func validateSourceKey(sourceSchema *entityops.Schema, key string, wantList bool) error {
	field, found := lo.Find(sourceSchema.Fields, func(f entityops.FieldDescriptor) bool {
		return f.InputKey == key
	})
	if !found {
		return fmt.Errorf("%w: %s is not a mapped input key on %s", ErrLinkSourceFieldInvalid, key, sourceSchema.Name)
	}

	if isList := strings.HasPrefix(field.Type, "[]"); isList != wantList {
		slot := "sourceField"
		if wantList {
			slot = "sourceList"
		}

		return fmt.Errorf("%w: %s.%s has type %s, which does not fit %s", ErrLinkSourceFieldInvalid, sourceSchema.Name, key, field.Type, slot)
	}

	return nil
}

// compileDefinition builds the indexed client, operation, and webhook event maps for one definition
func compileDefinition(def types.Definition) (definitionEntry, error) {
	credentialNames := indexCredentialNames(def.CredentialRegistrations)

	clients, err := indexClients(def.Clients, credentialNames)
	if err != nil {
		return definitionEntry{}, err
	}

	operations, err := indexOperations(def.Operations, clients)
	if err != nil {
		return definitionEntry{}, err
	}

	connections, err := indexConnections(def.Connections, credentialNames, clients, operations)
	if err != nil {
		return definitionEntry{}, err
	}

	webhooks, webhookEvents, err := indexWebhooks(def.Webhooks)
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

// indexCredentialNames builds a set of declared credential ref names for a definition
func indexCredentialNames(registrations []types.CredentialRegistration) map[string]struct{} {
	return lo.SliceToMap(registrations, func(reg types.CredentialRegistration) (string, struct{}) {
		return reg.Ref.String(), struct{}{}
	})
}

// indexClients indexes client registrations by client ref while validating credential cross-references
func indexClients(clients []types.ClientRegistration, credentialNames map[string]struct{}) (map[types.ClientID]types.ClientRegistration, error) {
	index := make(map[types.ClientID]types.ClientRegistration, len(clients))

	for _, client := range clients {
		if !client.Ref.Valid() {
			return nil, ErrClientRequired
		}

		for _, ref := range client.CredentialRefs {
			if _, declared := credentialNames[ref.String()]; !declared {
				return nil, ErrCredentialRefNotDeclared
			}
		}

		index[client.Ref] = client
	}

	return index, nil
}

// indexConnections indexes connection registrations while enforcing credential, client, and validation constraints
func indexConnections(connections []types.ConnectionRegistration, credentialNames map[string]struct{}, clients map[types.ClientID]types.ClientRegistration, operations map[string]types.OperationRegistration) (map[string]types.ConnectionRegistration, error) {
	connectionIndex := make(map[string]types.ConnectionRegistration, len(connections))

	for _, connection := range connections {
		if connection.CredentialRef == (types.CredentialSlotID{}) {
			return nil, ErrConnectionCredentialRefRequired
		}

		name := connection.CredentialRef.String()

		if _, declared := credentialNames[connection.CredentialRef.String()]; !declared {
			return nil, ErrConnectionCredentialRefNotDeclared
		}

		if !lo.Contains(connection.CredentialRefs, connection.CredentialRef) {
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

			if !lo.Contains(connection.CredentialRefs, connection.Auth.CredentialRef) {
				return nil, ErrConnectionAuthCredentialRefNotDeclared
			}
		}

		if connection.Disconnect != nil {
			if connection.Disconnect.CredentialRef == (types.CredentialSlotID{}) {
				return nil, ErrConnectionDisconnectCredentialRefNotDeclared
			}

			if !lo.Contains(connection.CredentialRefs, connection.Disconnect.CredentialRef) {
				return nil, ErrConnectionDisconnectCredentialRefNotDeclared
			}
		}

		connectionIndex[name] = connection
	}

	return connectionIndex, nil
}

// indexOperations indexes operations by name while validating handler and client cross-references
func indexOperations(operations []types.OperationRegistration, clients map[types.ClientID]types.ClientRegistration) (map[string]types.OperationRegistration, error) {
	index := make(map[string]types.OperationRegistration, len(operations))

	for _, operation := range operations {
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

		index[operation.Name] = operation
	}

	return index, nil
}

// indexWebhooks indexes webhook contracts and webhook events while validating structural constraints
func indexWebhooks(webhooks []types.WebhookRegistration) (map[string]types.WebhookRegistration, map[string]map[string]types.WebhookEventRegistration, error) {
	webhookIndex := make(map[string]types.WebhookRegistration, len(webhooks))
	webhookEventIndex := make(map[string]map[string]types.WebhookEventRegistration, len(webhooks))

	for _, webhook := range webhooks {
		if len(webhook.Events) > 0 && webhook.Event == nil {
			return nil, nil, ErrWebhookEventResolverRequired
		}

		eventIndex := make(map[string]types.WebhookEventRegistration, len(webhook.Events))

		for _, event := range webhook.Events {
			if event.Handle == nil {
				return nil, nil, ErrWebhookEventHandlerRequired
			}

			eventIndex[event.Name] = event
		}

		webhookIndex[webhook.Name] = webhook
		webhookEventIndex[webhook.Name] = eventIndex
	}

	return webhookIndex, webhookEventIndex, nil
}

// Listeners returns all operation registrations in stable topic order
func (r *Registry) Listeners() []types.OperationRegistration {
	return mapx.SortedValues(r.operationsByTopic, func(o types.OperationRegistration) gala.TopicName { return o.Topic })
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
	return mapx.SortedValues(r.webhookEventsByTopic, func(e types.WebhookEventRegistration) gala.TopicName { return e.Topic })
}

// GalaListeners returns all registered standalone gala listener registrations
func (r *Registry) GalaListeners() []types.GalaListenerRegistration {
	return append([]types.GalaListenerRegistration(nil), r.galaListeners...)
}

// RuntimeClient returns the cached runtime client for the given definition ID.
// Returns the client and true if a runtime integration was provisioned,
// or nil and false otherwise
func (r *Registry) RuntimeClient(definitionID string) (any, bool) {
	entry, ok := r.definitions[definitionID]
	if !ok || entry.runtimeClient == nil {
		return nil, false
	}

	return entry.runtimeClient, true
}

// StaticWebhooks returns all webhook registrations that declare a fixed static route
func (r *Registry) StaticWebhooks() []types.StaticWebhookEntry {
	var entries []types.StaticWebhookEntry

	for defID, entry := range r.definitions {
		for _, webhook := range entry.definition.Webhooks {
			if webhook.StaticRoute != "" {
				entries = append(entries, types.StaticWebhookEntry{
					DefinitionID: defID,
					WebhookName:  webhook.Name,
					StaticRoute:  webhook.StaticRoute,
				})
			}
		}
	}

	return entries
}

// IsRuntimeIntegration reports whether the given definition was provisioned
// as a runtime integration (no DB record, no keystore)
func (r *Registry) IsRuntimeIntegration(definitionID string) bool {
	entry, ok := r.definitions[definitionID]

	return ok && entry.definition.RuntimeIntegration != nil
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
