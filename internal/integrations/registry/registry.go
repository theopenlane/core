package registry

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// Builder builds one manifest-backed definition
type Builder func() (types.Definition, error)

// Registry is the in-memory index of registered definitions
type Registry struct {
	// definitions maps definition ID to its compiled entry
	definitions map[string]definitionEntry
	// operationsByTopic maps a topic name to its operation registration
	operationsByTopic map[gala.TopicName]types.OperationRegistration
	// webhookEventsByTopic maps a topic name to its webhook event registration
	webhookEventsByTopic map[gala.TopicName]types.WebhookEventRegistration
	// galaListeners collects standalone gala listener registrations across definitions
	galaListeners []types.GalaListenerRegistration
	// snapshots holds the committed surface snapshots every registered definition must match, nil when unchecked
	snapshots fs.FS
}

// Surface is a definition's installation-facing surface: every kind bound to stored data
type Surface struct {
	// ID is the canonical definition identifier
	ID string `json:"id"`
	// Credentials lists every credential slot with its stored schema, sorted by ref
	Credentials []SurfaceCredential `json:"credentials"`
	// UserInput is the stored user input schema when the definition declares a typed user input
	UserInput *SurfaceSchema `json:"userInput,omitempty"`
	// Installation is the derived installation metadata schema when the definition declares one
	Installation *SurfaceSchema `json:"installation,omitempty"`
	// Connections lists every connection mode's selecting credential ref, sorted
	Connections []string `json:"connections"`
	// Operations lists every operation with its retired names and config schema, sorted by name
	Operations []SurfaceOperation `json:"operations,omitempty"`
	// Webhooks lists every webhook contract with its events, sorted by name
	Webhooks []SurfaceWebhook `json:"webhooks,omitempty"`
}

// Snapshot is a definition's committed surface, its hash, and the version minted when the hash last changed
type Snapshot struct {
	// Hash is the SurfaceHash of the definition the snapshot was written from
	Hash string `json:"hash"`
	// Version is the ULID recorded on installations as their definition version
	Version string `json:"version"`
	// Surface is the definition's installation-facing surface
	Surface Surface `json:"surface"`
}

// SurfaceSchema is the stored schema of one kind and whether it declares an upgrade for older documents
type SurfaceSchema struct {
	// Schema is the reflected JSON schema of the stored type
	Schema json.RawMessage `json:"schema"`
	// Upgrade reports whether the kind declares an upgrade for documents stored under an older layout
	Upgrade bool `json:"upgrade,omitempty"`
}

// SurfaceCredential is one credential slot, the schema of what it stores, and the retired slots it takes over
type SurfaceCredential struct {
	// Ref is the stable credential slot name
	Ref string `json:"ref"`
	// SurfaceSchema is the stored credential schema with its upgrade declaration
	SurfaceSchema
	// Replaces lists the retired slot names whose stored payloads move onto this slot, sorted
	Replaces []string `json:"replaces,omitempty"`
}

// SurfaceOperation is one operation's name, retired names, and stored input schema
type SurfaceOperation struct {
	// Name is the stable operation name
	Name string `json:"name"`
	// Replaces lists the retired operation names this operation takes over, sorted
	Replaces []string `json:"replaces,omitempty"`
	// Schema is the composed JSON schema of the operation's stored input
	Schema json.RawMessage `json:"schema,omitempty"`
	// Upgrade reports whether the operation declares an upgrade for stored input persisted under an older layout
	Upgrade bool `json:"upgrade,omitempty"`
}

// SurfaceWebhook is one webhook contract with its events
type SurfaceWebhook struct {
	// Name is the stable webhook contract name
	Name string `json:"name"`
	// Replaces lists retired contract names whose webhook rows this contract takes over, sorted
	Replaces []string `json:"replaces,omitempty"`
	// Events lists the contract's event names, sorted
	Events []string `json:"events,omitempty"`
}

// DefinitionSurface projects a definition onto its installation-facing surface
func DefinitionSurface(def types.Definition) Surface {
	credentials := sortedProjection(def.CredentialRegistrations, func(registration types.CredentialRegistration) SurfaceCredential {
		replaces := lo.Map(registration.Replaces, func(slot types.CredentialSlotID, _ int) string { return slot.String() })

		return SurfaceCredential{Ref: registration.Ref.String(), SurfaceSchema: SurfaceSchema{Schema: registration.Stored.Schema, Upgrade: registration.Stored.Upgrade != nil}, Replaces: replaces}
	}, func(credential SurfaceCredential) string { return credential.Ref })

	connections := lo.Map(def.Connections, func(connection types.ConnectionRegistration, _ int) string {
		return connection.CredentialRef.String()
	})

	slices.Sort(connections)

	operations := sortedProjection(def.Operations, func(operation types.OperationRegistration) SurfaceOperation {
		surface := SurfaceOperation{Name: operation.Name, Replaces: operation.Replaces}

		if operation.Stored {
			surface.Schema = operation.Input.Schema
			surface.Upgrade = operation.Input.Upgrade != nil
		}

		return surface
	}, func(operation SurfaceOperation) string { return operation.Name })

	webhooks := sortedProjection(def.Webhooks, func(webhook types.WebhookRegistration) SurfaceWebhook {
		events := lo.Map(webhook.Events, func(event types.WebhookEventRegistration, _ int) string { return event.Name })

		slices.Sort(events)

		return SurfaceWebhook{Name: webhook.Name, Replaces: webhook.Replaces, Events: events}
	}, func(webhook SurfaceWebhook) string { return webhook.Name })

	surface := Surface{ID: def.ID, Credentials: credentials, Connections: connections, Operations: operations, Webhooks: webhooks}

	if def.UserInput != nil {
		surface.UserInput = &SurfaceSchema{Schema: def.UserInput.Schema, Upgrade: def.UserInput.Upgrade != nil}
	}

	if def.Installation != nil && len(def.Installation.Schema) > 0 {
		surface.Installation = &SurfaceSchema{Schema: def.Installation.Schema}
	}

	return surface
}

// sortedProjection projects each item and sorts the projections by key
func sortedProjection[T, R any](items []T, project func(T) R, key func(R) string) []R {
	out := lo.Map(items, func(item T, _ int) R { return project(item) })

	slices.SortFunc(out, func(a, b R) int {
		return strings.Compare(key(a), key(b))
	})

	return out
}

// definitionEntry captures the indexed details for one registered definition
type definitionEntry struct {
	// definition holds the original definition as supplied by the caller
	definition types.Definition
	// clients maps client ID to its client registration
	clients map[types.ClientID]types.ClientRegistration
	// operations maps operation name to its operation registration
	operations map[string]types.OperationRegistration
	// webhooks maps webhook name to its webhook registration
	webhooks map[string]types.WebhookRegistration
	// webhookEvents maps webhook name to a nested map of event name to event registration
	webhookEvents map[string]map[string]types.WebhookEventRegistration
	// runtimeClient holds the pre-built client for runtime integrations
	runtimeClient any
	// version is the ULID of the committed snapshot the definition matched, empty without snapshots
	version string
}

// Option configures a Registry
type Option func(*Registry)

// WithSnapshots requires every registered definition to match its committed snapshot in fsys and records the snapshot's version
func WithSnapshots(fsys fs.FS) Option {
	return func(r *Registry) {
		r.snapshots = fsys
	}
}

// New constructs an empty registry
func New(opts ...Option) *Registry {
	r := &Registry{
		definitions:          map[string]definitionEntry{},
		operationsByTopic:    map[gala.TopicName]types.OperationRegistration{},
		webhookEventsByTopic: map[gala.TopicName]types.WebhookEventRegistration{},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Register adds one definition to the registry
func (r *Registry) Register(def types.Definition) error {
	def, err := finalizeDefinition(def)
	if err != nil {
		return err
	}

	if err := r.validateDefinition(def); err != nil {
		return err
	}

	populateMappingLinkTargets(def.Mappings)

	entry, err := compileDefinition(def)
	if err != nil {
		return err
	}

	if r.snapshots != nil {
		entry.version, err = committedVersion(r.snapshots, def)
		if err != nil {
			return err
		}
	}

	if def.RuntimeIntegration != nil && def.RuntimeIntegration.Config != nil {
		client, buildErr := def.RuntimeIntegration.Build(context.Background(), def.RuntimeIntegration.Config)
		if buildErr != nil {
			return buildErr
		}

		entry.runtimeClient = client
	}

	operationTopics, err := indexUnique(def.ID, "operation topic", def.Operations, func(operation types.OperationRegistration) gala.TopicName {
		return operation.Topic
	}, r.operationsByTopic)
	if err != nil {
		return err
	}

	events := lo.FlatMap(def.Webhooks, func(webhook types.WebhookRegistration, _ int) []types.WebhookEventRegistration { return webhook.Events })

	webhookEventTopics, err := indexUnique(def.ID, "webhook event topic", events, func(event types.WebhookEventRegistration) gala.TopicName {
		return event.Topic
	}, r.webhookEventsByTopic)
	if err != nil {
		return err
	}

	maps.Copy(r.operationsByTopic, operationTopics)
	maps.Copy(r.webhookEventsByTopic, webhookEventTopics)

	r.definitions[def.ID] = entry

	r.galaListeners = append(r.galaListeners, def.GalaListeners...)

	return nil
}

// indexUnique indexes items by key, rejecting a key repeated within items or already held by taken
func indexUnique[K comparable, V any](definitionID, kind string, items []V, key func(V) K, taken map[K]V) (map[K]V, error) {
	index := make(map[K]V, len(items))

	for _, item := range items {
		name := key(item)
		if lo.HasKey(index, name) || lo.HasKey(taken, name) {
			return nil, fmt.Errorf("%w: definition %s %s %q", ErrDuplicateRegistration, definitionID, kind, fmt.Sprint(name))
		}

		index[name] = item
	}

	return index, nil
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

	return entry.definition, ok
}

// Version returns the committed snapshot version of one definition, or empty when unregistered or registered without snapshots
func (r *Registry) Version(id string) string {
	return r.definitions[id].version
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
	switch {
	case def.ID == "":
		return ErrDefinitionIDRequired
	case lo.HasKey(r.definitions, def.ID):
		return ErrDefinitionAlreadyRegistered
	case def.OperatorConfig != nil && len(def.OperatorConfig.Schema) == 0:
		return ErrOperatorConfigSchemaRequired
	case lo.ContainsBy(def.CredentialRegistrations, func(credential types.CredentialRegistration) bool { return len(credential.Stored.Schema) == 0 }):
		return ErrCredentialSchemaRequired
	case def.UserInput != nil && len(def.UserInput.Schema) == 0:
		return ErrUserInputSchemaRequired
	case def.RuntimeIntegration != nil && def.RuntimeIntegration.Build == nil:
		return ErrRuntimeBuildRequired
	}

	if err := validateHealthCheck(def); err != nil {
		return err
	}

	return validateMappingLinks(def.Mappings)
}

// validateHealthCheck requires a health check handler when connections are declared
func validateHealthCheck(def types.Definition) error {
	check := def.HealthCheck

	switch {
	case check == nil && len(def.Connections) > 0:
		return ErrHealthCheckRequired
	case check == nil:
		return nil
	case check.Handle == nil:
		return ErrConnectionHealthCheckHandlerRequired
	case !check.ClientRef.Valid():
		return nil
	}

	client, declared := lo.Find(def.Clients, func(client types.ClientRegistration) bool {
		return client.Ref == check.ClientRef
	})
	if !declared {
		return ErrConnectionClientRefNotDeclared
	}

	uncovered, found := lo.Find(def.Connections, func(connection types.ConnectionRegistration) bool {
		return !lo.Contains(client.CredentialRefs, connection.CredentialRef)
	})
	if found {
		return fmt.Errorf("%w: definition %s client %s connection %s", ErrHealthCheckClientCredentialMissing, def.ID, check.ClientRef, uncovered.CredentialRef)
	}

	return nil
}

// populateMappingLinkTargets fills each mapping's cross-link inventory from the entityops catalog
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

// targetLinkFields projects a target schema's match-key fields
func targetLinkFields(fields []entityops.FieldDescriptor) []types.LinkFieldInfo {
	return lo.FilterMap(fields, func(f entityops.FieldDescriptor, _ int) (types.LinkFieldInfo, bool) {
		if !f.MatchKey {
			return types.LinkFieldInfo{}, false
		}

		return types.LinkFieldInfo{Name: f.Name, Label: f.Label, Type: f.Type}, true
	})
}

// sourceLinkFields projects the source schema's mapped input keys
func sourceLinkFields(fields []entityops.FieldDescriptor) []types.LinkFieldInfo {
	return lo.FilterMap(fields, func(f entityops.FieldDescriptor, _ int) (types.LinkFieldInfo, bool) {
		if f.InputKey == "" {
			return types.LinkFieldInfo{}, false
		}

		return types.LinkFieldInfo{Name: f.InputKey, Label: f.Label, Type: f.Type}, true
	})
}

// ResolveLinkEdge resolves the edge a link rule links through, by name or unique target type
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

// validateMappingLinks validates every link rule a mapping declares against the entityops catalog
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

// ValidateLinkRules validates each rule's edge, match shape, and referenced fields
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

// validateLinkRuleFields checks one resolved rule's match configuration
func validateLinkRuleFields(sourceSchema *entityops.Schema, edge entityops.EdgeDescriptor, rule types.LinkRule) error {
	if edge.Target == nil {
		return fmt.Errorf("%w: %s.%s targets %s", ErrLinkTargetNotRegistered, sourceSchema.Name, edge.Name, edge.TargetType)
	}

	fieldMatch := rule.TargetField != "" && (rule.SourceField != "" || rule.SourceList != "")

	if fieldMatch == (rule.Expression != "") {
		return fmt.Errorf("%w: %s -> %s must set either a target/source field match or an expression", ErrLinkRuleInvalid, sourceSchema.Name, edge.Name)
	}

	if !fieldMatch {
		return nil
	}

	if !edge.Target.MatchKeyField(rule.TargetField) {
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

// validateSourceKey checks key is a mapped input key with the matching scalar/list shape
func validateSourceKey(sourceSchema *entityops.Schema, key string, wantList bool) error {
	field, found := lo.Find(sourceSchema.Fields, func(f entityops.FieldDescriptor) bool {
		return f.InputKey == key
	})
	if !found {
		return fmt.Errorf("%w: %s is not a mapped input key on %s", ErrLinkSourceFieldInvalid, key, sourceSchema.Name)
	}

	if isList := strings.HasPrefix(field.Type, "[]"); isList != wantList {
		return fmt.Errorf("%w: %s.%s has type %s, which does not fit %s", ErrLinkSourceFieldInvalid, sourceSchema.Name, key, field.Type, lo.Ternary(wantList, "sourceList", "sourceField"))
	}

	return nil
}

// compileDefinition builds the indexed client, operation, and webhook event maps for one definition
func compileDefinition(def types.Definition) (definitionEntry, error) {
	declared := lo.Map(def.CredentialRegistrations, func(registration types.CredentialRegistration, _ int) types.CredentialSlotID {
		return registration.Ref
	})

	clients, err := indexClients(def.ID, def.Clients, declared)
	if err != nil {
		return definitionEntry{}, err
	}

	operations, err := indexOperations(def.ID, def.Operations, clients)
	if err != nil {
		return definitionEntry{}, err
	}

	if err := validateConnections(def.ID, def.Connections, declared); err != nil {
		return definitionEntry{}, err
	}

	webhooks, webhookEvents, err := indexWebhooks(def.ID, def.Webhooks)
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

// indexClients indexes clients by ref, requiring declared credential slots
func indexClients(definitionID string, clients []types.ClientRegistration, declared []types.CredentialSlotID) (map[types.ClientID]types.ClientRegistration, error) {
	for _, client := range clients {
		switch {
		case !client.Ref.Valid():
			return nil, ErrClientRequired
		case !lo.Every(declared, client.CredentialRefs):
			return nil, ErrCredentialRefNotDeclared
		}
	}

	return indexUnique(definitionID, "client", clients, func(client types.ClientRegistration) types.ClientID { return client.Ref }, nil)
}

// validateConnections requires each connection's slots to be declared and its selecting slot unique
func validateConnections(definitionID string, connections []types.ConnectionRegistration, declared []types.CredentialSlotID) error {
	for _, connection := range connections {
		slots := append([]types.CredentialSlotID{connection.CredentialRef}, connection.CredentialRefs...)

		switch {
		case connection.CredentialRef == (types.CredentialSlotID{}):
			return ErrConnectionCredentialRefRequired
		case !lo.Every(declared, slots):
			return ErrConnectionCredentialRefNotDeclared
		case connection.Auth != nil && !lo.Contains(slots, connection.Auth.CredentialRef):
			return ErrConnectionAuthCredentialRefNotDeclared
		case connection.Disconnect != nil && !lo.Contains(slots, connection.Disconnect.CredentialRef):
			return ErrConnectionDisconnectCredentialRefNotDeclared
		}
	}

	_, err := indexUnique(definitionID, "connection", connections, func(connection types.ConnectionRegistration) types.CredentialSlotID {
		return connection.CredentialRef
	}, nil)

	return err
}

// indexOperations indexes operations by name, validating handler and client requirements
func indexOperations(definitionID string, operations []types.OperationRegistration, clients map[types.ClientID]types.ClientRegistration) (map[string]types.OperationRegistration, error) {
	for _, operation := range operations {
		switch {
		case operation.Handle == nil && operation.IngestHandle == nil:
			return nil, ErrOperationHandlerRequired
		case operation.Handle != nil && operation.IngestHandle != nil:
			return nil, ErrOperationHandlerAmbiguous
		case operation.IngestHandle != nil && len(operation.Ingest) == 0:
			return nil, ErrIngestContractsRequired
		case operation.Policy.Snapshot && operation.IngestHandle == nil:
			return nil, ErrIngestSnapshotRequiresIngestHandle
		case operation.ClientRef.Valid() && !lo.HasKey(clients, operation.ClientRef):
			return nil, ErrClientNotFound
		}
	}

	return indexUnique(definitionID, "operation", operations, func(operation types.OperationRegistration) string { return operation.Name }, nil)
}

// indexWebhooks indexes webhook contracts and events, requiring resolver and handlers
func indexWebhooks(definitionID string, webhooks []types.WebhookRegistration) (map[string]types.WebhookRegistration, map[string]map[string]types.WebhookEventRegistration, error) {
	webhookEvents := make(map[string]map[string]types.WebhookEventRegistration, len(webhooks))

	for _, webhook := range webhooks {
		switch {
		case len(webhook.Events) > 0 && webhook.Event == nil:
			return nil, nil, ErrWebhookEventResolverRequired
		case lo.ContainsBy(webhook.Events, func(event types.WebhookEventRegistration) bool { return event.Handle == nil }):
			return nil, nil, ErrWebhookEventHandlerRequired
		}

		events, err := indexUnique(definitionID, "webhook event", webhook.Events, func(event types.WebhookEventRegistration) string { return event.Name }, nil)
		if err != nil {
			return nil, nil, err
		}

		webhookEvents[webhook.Name] = events
	}

	index, err := indexUnique(definitionID, "webhook", webhooks, func(webhook types.WebhookRegistration) string { return webhook.Name }, nil)
	if err != nil {
		return nil, nil, err
	}

	return index, webhookEvents, nil
}

// Listeners returns all operation registrations in stable topic order
func (r *Registry) Listeners() []types.OperationRegistration {
	return mapx.SortedValues(r.operationsByTopic, func(o types.OperationRegistration) gala.TopicName { return o.Topic })
}

// WebhookEvent returns one webhook event registration for a definition
func (r *Registry) WebhookEvent(id string, webhookName string, eventName string) (types.WebhookEventRegistration, error) {
	events, err := lookupInEntry(r, id, webhookName, func(e definitionEntry) map[string]map[string]types.WebhookEventRegistration {
		return e.webhookEvents
	}, ErrWebhookNotFound)
	if err != nil {
		return types.WebhookEventRegistration{}, err
	}

	event, ok := events[eventName]
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

// RuntimeClient returns the cached runtime client for the given definition ID
func (r *Registry) RuntimeClient(definitionID string) (any, bool) {
	client := r.definitions[definitionID].runtimeClient

	return client, client != nil
}

// StaticWebhooks returns all webhook registrations that declare a fixed static route, sorted by definition id then webhook name
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

	slices.SortFunc(entries, func(a, b types.StaticWebhookEntry) int {
		return cmp.Or(strings.Compare(a.DefinitionID, b.DefinitionID), strings.Compare(a.WebhookName, b.WebhookName))
	})

	return entries
}

// IsRuntimeIntegration reports whether the definition was provisioned as a runtime integration
func (r *Registry) IsRuntimeIntegration(definitionID string) bool {
	return r.definitions[definitionID].definition.RuntimeIntegration != nil
}

// lookupInEntry finds a definition entry, then looks up a value in its sub-map
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
